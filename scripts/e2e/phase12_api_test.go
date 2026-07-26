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
	p12ApiURL = "http://localhost:8090"
	p12DbURL  = "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
)

func waitForP12Services(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 5; i++ {
		resp, err := client.Get(p12ApiURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("API server not reachable at %s. Please ensure the local server is running.", p12ApiURL)
}

func setupP12Database(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, p12DbURL)
	if err != nil {
		t.Fatal("Failed to connect to PGlite Database")
	}

	_, err = pool.Exec(ctx, "DELETE FROM dependencies"); require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM contracts"); require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM repositories"); require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM organizations"); require.NoError(t, err)

	// Seed data
	_, err = pool.Exec(ctx, `INSERT INTO organizations (id, github_installation_id, github_org_name) VALUES ('11111111-1111-1111-1111-111111111111', 123456, 'mcp-org') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO repositories (id, org_id, github_repo_id, name, full_name) VALUES ('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 101, 'backend', 'mcp-org/backend'), ('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 102, 'frontend', 'mcp-org/frontend'), ('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 103, 'core-repo', 'mcp-org/core-repo') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO contracts (id, repo_id, schema_type, spec_path, branch, latest_commit_sha, raw_content) VALUES ('55555555-5555-5555-5555-555555555555', '22222222-2222-2222-2222-222222222222', 'openapi', 'openapi.yaml', 'main', 'sha-backend', 'openapi: 3.0.0') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	_, err = pool.Exec(ctx, `INSERT INTO dependencies (id, consumer_repo_id, provider_contract_id) VALUES ('66666666-6666-6666-6666-666666666666', '33333333-3333-3333-3333-333333333333', '55555555-5555-5555-5555-555555555555') ON CONFLICT DO NOTHING`)
	require.NoError(t, err)
	return pool
}

func TestPhase12API_GraphEndpoint(t *testing.T) {
	waitForP12Services(t)

	pool := setupP12Database(t)
	defer pool.Close()

	ctx := context.Background()
	var orgName string
	err := pool.QueryRow(ctx, "SELECT github_org_name FROM organizations WHERE github_org_name = 'mcp-org' LIMIT 1").Scan(&orgName)
	require.NoError(t, err, "Pre-seeded organization 'mcp-org' should exist in the database")
	assert.Equal(t, "mcp-org", orgName)

	req, err := http.NewRequest("GET", p12ApiURL+"/api/v1/graph/mcp-org", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer local-dev-token")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	foundEdge := false
	for _, edge := range result {
		consumer, _ := edge["consumer"].(string)
		provider, _ := edge["provider"].(string)
		if consumer == "mcp-org/frontend" && provider == "mcp-org/backend" {
			foundEdge = true
			break
		}
	}
	assert.True(t, foundEdge, "Expected graph to contain edge from mcp-org/frontend to mcp-org/backend")
}
