package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	p3ApiURL           = "http://localhost:8090"
	p3DbURL            = "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
	p3RegistryAPIToken = "local-dev-token"
)

func waitForP3Services(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(p3ApiURL + "/health")
	if err != nil || resp.StatusCode != 200 {
		t.Skipf("API server not reachable at %s. Skipping E2E test.", p3ApiURL)
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func setupP3Database(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, p3DbURL)
	if err != nil {
		t.Skip("Failed to connect to real PostgreSQL. Skipping E2E test.")
	}

	tables := []string{
		"drift_anomalies",
		"endpoint_traffic",
		"governance_rules",
		"breaking_change_history",
		"preview_sessions", "diff_reports",
		"dependencies",
		"contracts",
		"repositories",
		"organizations",
	}

	for _, table := range tables {
		_, err = pool.Exec(ctx, fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Failed to delete %s: %v", table, err)
		}
	}

	return pool
}

func TestPhase3ContractRegistry(t *testing.T) {
	waitForP3Services(t)
	pool := setupP3Database(t)
	defer pool.Close()

	ctx := context.Background()

	// Seed database
	var orgID string
	err := pool.QueryRow(ctx, "INSERT INTO organizations (github_installation_id, github_org_name) VALUES ($1, $2) RETURNING id", 3001, "p3-org").Scan(&orgID)
	require.NoError(t, err)

	var providerRepoID string
	err = pool.QueryRow(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name, metadata) VALUES ($1, $2, $3, $4, $5) RETURNING id", orgID, 3101, "backend", "p3-org/backend", "{}").Scan(&providerRepoID)
	require.NoError(t, err)

	var consumerRepoID string
	err = pool.QueryRow(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name, metadata) VALUES ($1, $2, $3, $4, $5) RETURNING id", orgID, 3102, "frontend", "p3-org/frontend", "{}").Scan(&consumerRepoID)
	require.NoError(t, err)

	var contractID string
	err = pool.QueryRow(ctx, "INSERT INTO contracts (repo_id, schema_type, spec_path, latest_commit_sha, raw_content) VALUES ($1, $2, $3, $4, $5) RETURNING id", providerRepoID, "openapi", "openapi.yaml", "sha-base", "mock-content").Scan(&contractID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO dependencies (consumer_repo_id, provider_contract_id, status) VALUES ($1, $2, $3)", consumerRepoID, contractID, "active")
	require.NoError(t, err)

	t.Run("Scenario 1: Dependency graph returns correct edges", func(t *testing.T) {
		req, _ := http.NewRequest("GET", p3ApiURL+"/api/v1/graph/p3-org", nil)
		req.Header.Set("Authorization", "Bearer "+p3RegistryAPIToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var edges []map[string]interface{}
		body, _ := io.ReadAll(resp.Body)
		err = json.Unmarshal(body, &edges)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(edges), 1)

		found := false
		for _, edge := range edges {
			if edge["consumer"] == "p3-org/frontend" {
				found = true
				assert.Equal(t, "p3-org/backend", edge["provider"])
				assert.Equal(t, "active", edge["status"])
			}
		}
		assert.True(t, found, "Expected to find edge with consumer p3-org/frontend")
	})

	t.Run("Scenario 2: Can-deploy gate blocks breaking change", func(t *testing.T) {
		_, err = pool.Exec(ctx, "INSERT INTO breaking_change_history (repo_id, org_name, repo_name, git_sha, breaking_changes) VALUES ($1, $2, $3, $4, $5)", providerRepoID, "p3-org", "backend", "sha-breaking", `{"breaking_count":1}`)
		require.NoError(t, err)

		req, _ := http.NewRequest("GET", p3ApiURL+"/api/v1/registry/can-deploy?repo=p3-org/backend&commit=sha-breaking", nil)
		req.Header.Set("Authorization", "Bearer "+p3RegistryAPIToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(body), "p3-org/frontend")
	})

	t.Run("Scenario 3: Can-deploy gate passes safe commit", func(t *testing.T) {
		req, _ := http.NewRequest("GET", p3ApiURL+"/api/v1/registry/can-deploy?repo=p3-org/backend&commit=sha-safe", nil)
		req.Header.Set("Authorization", "Bearer "+p3RegistryAPIToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("Scenario 4: List repos returns registered repos", func(t *testing.T) {
		req, _ := http.NewRequest("GET", p3ApiURL+"/api/v1/repos/p3-org", nil)
		req.Header.Set("Authorization", "Bearer "+p3RegistryAPIToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		body, _ := io.ReadAll(resp.Body)

		// The handler returns a slice of db.Repository but the struct is likely nested/marshaled.
		// We'll unmarshal to a generic slice and verify properties or simple string matches
		// Just unmarshal as generic interface and check len, and use string match for robustness.
		var repos []map[string]interface{}
		err = json.Unmarshal(body, &repos)
		require.NoError(t, err)

		assert.GreaterOrEqual(t, len(repos), 2)

		bodyStr := string(body)
		assert.Contains(t, bodyStr, "p3-org/backend")
		assert.Contains(t, bodyStr, "p3-org/frontend")
	})
}
