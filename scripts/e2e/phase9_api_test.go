package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	p9ApiURL = "http://localhost:8090"
	p9DbURL  = "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
)

func p9WaitForServices(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 5; i++ {
		resp, err := client.Get(p9ApiURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	t.Skipf("API server not reachable at %s. Skipping E2E test.", p9ApiURL)
}

func p9SetupDatabase(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, p9DbURL)
	require.NoError(t, err, "Failed to connect to PGlite PostgreSQL")

	// Clean tables at setup to leave data intact for Playwright
	_, err = pool.Exec(ctx, "DELETE FROM dependencies")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM contracts")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM repositories")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM organizations")
	require.NoError(t, err)

	// Seed data for Phase 9
	var orgID string
	err = pool.QueryRow(ctx, "INSERT INTO organizations (id, github_installation_id, github_org_name) VALUES ('11111111-1111-1111-1111-111111111111', 123456, 'mcp-org') RETURNING id").Scan(&orgID)
	require.NoError(t, err)

	// Core Repo for generic list
	_, err = pool.Exec(ctx, "INSERT INTO repositories (id, org_id, github_repo_id, name, full_name) VALUES ('22222222-2222-2222-2222-222222222222', $1, 1234, 'core-repo', 'mcp-org/core-repo')", orgID)
	require.NoError(t, err)

	// Graph test repos (provider and consumer)
	_, err = pool.Exec(ctx, "INSERT INTO repositories (id, org_id, github_repo_id, name, full_name) VALUES ('33333333-3333-3333-3333-333333333333', $1, 101, 'core-consumer', 'mcp-org/core-consumer')", orgID)
	require.NoError(t, err)

	// Add contract to core-repo
	_, err = pool.Exec(ctx, "INSERT INTO contracts (id, repo_id, schema_type, spec_path, branch, latest_commit_sha, raw_content) VALUES ('55555555-5555-5555-5555-555555555555', '22222222-2222-2222-2222-222222222222', 'openapi', 'openapi.yaml', 'main', 'sha-backend', 'openapi: 3.0.0')")
	require.NoError(t, err)

	// Add dependency edge from core-consumer -> core-repo
	_, err = pool.Exec(ctx, "INSERT INTO dependencies (id, consumer_repo_id, provider_contract_id) VALUES ('66666666-6666-6666-6666-666666666666', '33333333-3333-3333-3333-333333333333', '55555555-5555-5555-5555-555555555555')")
	require.NoError(t, err)

	return pool
}

func TestPhase9ComplianceAndRiskScoreE2E(t *testing.T) {
	p9WaitForServices(t)

	pool := p9SetupDatabase(t)
	// We deliberately DO NOT defer pool.Exec("DELETE FROM...") to allow Playwright tests to use the seeded state!
	defer pool.Close()

	t.Run("Scenario 1: Risk Score API - LOW", func(t *testing.T) {
		req, err := http.NewRequest("GET", p9ApiURL+"/api/v1/risk/mcp-org/core-repo/101?e2e_tests_pass=true", nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "LOW", result["score"])
	})

	t.Run("Scenario 2: Risk Score API - HIGH", func(t *testing.T) {
		req, err := http.NewRequest("GET", p9ApiURL+"/api/v1/risk/mcp-org/core-repo/102?has_db_migrations=true&e2e_tests_pass=true", nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "HIGH", result["score"])
	})

	t.Run("Scenario 3: Risk Score API - CRITICAL", func(t *testing.T) {
		req, err := http.NewRequest("GET", p9ApiURL+"/api/v1/risk/mcp-org/core-repo/103?has_db_migrations=true&e2e_tests_pass=false", nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, "CRITICAL", result["score"])
	})

	t.Run("Scenario 4: Risk Score API - Missing Params", func(t *testing.T) {
		// Chi router will return 404 if path params are missing
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/risk///123", p9ApiURL), nil)
		require.NoError(t, err)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Chi Router handles this as not found since empty path params don't match the route
		assert.Contains(t, []int{http.StatusNotFound, http.StatusBadRequest, http.StatusMovedPermanently}, resp.StatusCode)
	})
}
