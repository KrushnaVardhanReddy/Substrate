package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	p4ApiURL           = "http://localhost:8090"
	p4DbURL            = "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
	p4RegistryAPIToken = "local-dev-token"
)

func waitForP4Services(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(p4ApiURL + "/health")
	if err != nil || resp.StatusCode != 200 {
		t.Skipf("API server not reachable at %s. Skipping E2E test.", p4ApiURL)
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func setupP4Database(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, p4DbURL)
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
		_, _ = pool.Exec(ctx, fmt.Sprintf("DELETE FROM %s", table))
	}
	return pool
}

// P4-T10 — Phase 4 AI Diff Engine Full-Stack E2E Validation
func TestPhase4AIDiff(t *testing.T) {
	if (os.Getenv("LLM_API_KEY") == "" || os.Getenv("LLM_API_KEY") == "dummy") && os.Getenv("OPENROUTER_API_KEY") == "" {
		t.Skip("Skipping real LLM test because LLM_API_KEY is missing or dummy")
	}

	waitForP4Services(t)
	pool := setupP4Database(t)
	defer pool.Close()

	ctx := context.Background()

	var orgID string
	err := pool.QueryRow(ctx, "INSERT INTO organizations (github_installation_id, github_org_name) VALUES ($1, $2) RETURNING id", 4001, "mcp-org").Scan(&orgID)
	require.NoError(t, err)

	var providerRepoID string
	err = pool.QueryRow(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name, metadata) VALUES ($1, $2, $3, $4, $5) RETURNING id", orgID, 4101, "test-repo", "mcp-org/test-repo", "{}").Scan(&providerRepoID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO contracts (repo_id, schema_type, spec_path, latest_commit_sha, raw_content)
		VALUES ($1, 'openapi', 'openapi.yaml', 'test-sha', 'openapi: 3.0.0')
	`, providerRepoID)
	require.NoError(t, err)

	t.Run("POST /api/v1/ai/analyze (SSE stream)", func(t *testing.T) {
		reqData := map[string]interface{}{
			"org":             "mcp-org",
			"schema_type":     "openapi",
			"current_schema":  "openapi: 3.0.0\ninfo:\n  title: Test API\n  version: 1.0.0\npaths:\n  /test:\n    get:\n      responses:\n        '200':\n          description: OK",
			"proposed_schema": "openapi: 3.0.0\ninfo:\n  title: Test API\n  version: 1.0.0\npaths:\n  /test:\n    post:\n      responses:\n        '200':\n          description: OK",
		}

		body, _ := json.Marshal(reqData)
		req, err := http.NewRequest("POST", p4ApiURL+"/api/v1/ai/analyze", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+p4RegistryAPIToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

		bodyBytes, _ := io.ReadAll(resp.Body)
		bodyStr := string(bodyBytes)

		// Check for the "done" event in the mock stream output
		if !strings.Contains(bodyStr, "data: {\"type\":\"done\"}") && !strings.Contains(bodyStr, "data:{\"type\":\"done\"}") {
			t.Errorf("expected done event in SSE stream, got: %v", bodyStr)
		}
	})

	t.Run("POST /api/v1/ai/autofix", func(t *testing.T) {
		reqData := map[string]interface{}{
			"org":           "mcp-org",
			"provider_repo": "mcp-org/test-repo",
		}

		body, _ := json.Marshal(reqData)
		req, err := http.NewRequest("POST", p4ApiURL+"/api/v1/ai/autofix", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+p4RegistryAPIToken)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var res map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&res)
		require.NoError(t, err)

		explanation, ok := res["explanation"].(string)
		assert.True(t, ok, "expected explanation string in response")
		assert.NotEmpty(t, explanation, "expected explanation to not be empty")
	})
}
