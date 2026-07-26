package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	p11UIAPIURL   = "http://localhost:8090"
	p11UIDBURL    = "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
	p11UIAPIToken = "local-dev-token"
)

func waitForP11UIServices(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(p11UIAPIURL + "/health")
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("API server not reachable at %s. Failing E2E test.", p11UIAPIURL)
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func setupP11UIDatabase(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, p11UIDBURL)
	if err != nil {
		t.Fatal("Failed to connect to PGlite PostgreSQL. Failing E2E test.")
	}

	_, err = pool.Exec(ctx, "DELETE FROM dependencies")
	if err != nil {
		t.Logf("Failed to delete dependencies: %v", err)
	}
	_, err = pool.Exec(ctx, "DELETE FROM contracts")
	if err != nil {
		t.Logf("Failed to delete contracts: %v", err)
	}
	_, err = pool.Exec(ctx, "DELETE FROM repositories")
	if err != nil {
		t.Logf("Failed to delete repositories: %v", err)
	}
	_, err = pool.Exec(ctx, "DELETE FROM organizations")
	if err != nil {
		t.Logf("Failed to delete organizations: %v", err)
	}

	// Seed db
	_, err = pool.Exec(ctx, `
		INSERT INTO organizations (id, github_installation_id, github_org_name) VALUES
		('11111111-1111-1111-1111-111111111111', 123456, 'mcp-org')
		ON CONFLICT (id) DO NOTHING;

		INSERT INTO repositories (id, org_id, github_repo_id, name, full_name) VALUES
		('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 101, 'backend', 'mcp-org/backend'),
		('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 102, 'frontend', 'mcp-org/frontend'),
		('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 103, 'core-repo', 'mcp-org/core-repo')
		ON CONFLICT (id) DO NOTHING;

		INSERT INTO contracts (id, repo_id, schema_type, spec_path, branch, latest_commit_sha, raw_content) VALUES
		('55555555-5555-5555-5555-555555555555', '22222222-2222-2222-2222-222222222222', 'openapi', 'openapi.yaml', 'main', 'sha-backend', 'openapi: 3.0.0')
		ON CONFLICT (id) DO NOTHING;

		INSERT INTO dependencies (id, consumer_repo_id, provider_contract_id) VALUES
		('66666666-6666-6666-6666-666666666666', '33333333-3333-3333-3333-333333333333', '55555555-5555-5555-5555-555555555555')
		ON CONFLICT (id) DO NOTHING;
	`)
	require.NoError(t, err)

	return pool
}

func TestPhase11API(t *testing.T) {
	waitForP11UIServices(t)

	pool := setupP11UIDatabase(t)
	defer pool.Close()

	t.Run("Scenario 1: Verify Dependency Graph returns correct edges for mcp-org", func(t *testing.T) {
		req, err := http.NewRequest("GET", p11UIAPIURL+"/api/v1/graph/mcp-org", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+p11UIAPIToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		require.Greater(t, len(result), 0)

		// frontend depends on backend
		found := false
		for _, edge := range result {
			if edge["consumer"] == "mcp-org/frontend" && edge["provider"] == "mcp-org/backend" {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected to find an edge where consumer is mcp-org/frontend and provider is mcp-org/backend")
	})
}
