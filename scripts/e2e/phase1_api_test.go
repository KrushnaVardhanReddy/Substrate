package main

import (
	"bytes"
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
	p1ApiURL           = "http://localhost:8090"
	p1DbURL            = "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
	p1RegistryAPIToken = "local-dev-token"
)

func waitForP1Services(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(p1ApiURL + "/health")
	if err != nil || resp.StatusCode != 200 {
		t.Fatalf("API server not reachable at %s.", p1ApiURL)
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func setupP1Database(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, p1DbURL)
	require.NoError(t, err, "Failed to connect to real PostgreSQL")

	_, _ = pool.Exec(ctx, "DELETE FROM dependencies")
	_, _ = pool.Exec(ctx, "DELETE FROM contracts")
	_, _ = pool.Exec(ctx, "DELETE FROM repositories")
	_, _ = pool.Exec(ctx, "DELETE FROM organizations")

	return pool
}

func TestPhase1CoreDiffEngine(t *testing.T) {
	waitForP1Services(t)
	pool := setupP1Database(t)
	defer pool.Close()

	ctx := context.Background()

	var orgID string
	err := pool.QueryRow(ctx, "INSERT INTO organizations (id, github_installation_id, github_org_name) VALUES ('11111111-1111-1111-1111-111111111111', 123456, 'mcp-org') ON CONFLICT (id) DO UPDATE SET github_installation_id = EXCLUDED.github_installation_id RETURNING id").Scan(&orgID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO repositories (id, org_id, github_repo_id, name, full_name) VALUES ('22222222-2222-2222-2222-222222222222', $1, 201, 'core-repo', 'mcp-org/core-repo') ON CONFLICT DO NOTHING", orgID)
	require.NoError(t, err)

	t.Run("Push Baseline OpenAPI Spec via diff", func(t *testing.T) {
		diffReq := map[string]interface{}{
			"diff_report": map[string]interface{}{
				"summary": map[string]interface{}{
					"breaking_count": 0,
					"safe_count":     0,
				},
				"schema_type": "openapi",
				"version":     "v1",
			},
			"is_audit_mode":       false,
			"org":                 "mcp-org",
			"provider_repo":       "mcp-org/core-repo",
			"pr_number":           1,
			"commit_sha":          "base-sha",
			"head_schema_content": "openapi: \"3.0.0\"\ninfo:\n  title: Provider API\n  version: \"1.0.0\"\npaths:\n  /users:\n    get:\n      summary: List users\n      responses:\n        \"200\":\n          description: OK",
			"schema_type":         "openapi",
		}

		diffBody, _ := json.Marshal(diffReq)
		reqDiff, err := http.NewRequest("POST", p1ApiURL+"/api/v1/diff", bytes.NewReader(diffBody))
		require.NoError(t, err)
		reqDiff.Header.Set("Content-Type", "application/json")
		reqDiff.Header.Set("Authorization", "Bearer "+p1RegistryAPIToken)

		respDiff, err := http.DefaultClient.Do(reqDiff)
		require.NoError(t, err)
		defer respDiff.Body.Close()

		assert.Equal(t, http.StatusCreated, respDiff.StatusCode)
	})

	t.Run("Push breaking change (removed required field / endpoint)", func(t *testing.T) {
		diffReq := map[string]interface{}{
			"diff_report": map[string]interface{}{
				"summary": map[string]interface{}{
					"breaking_count": 1,
					"safe_count":     0,
				},
				"breaking_changes": []map[string]interface{}{
					{
						"rule_id": "ENDPOINT_REMOVED",
						"message": "Endpoint /users was removed.",
					},
				},
				"schema_type": "openapi",
				"version":     "v1",
			},
			"is_audit_mode":       false,
			"org":                 "mcp-org",
			"provider_repo":       "mcp-org/core-repo",
			"pr_number":           2,
			"commit_sha":          "breaking-sha",
			"head_schema_content": "openapi: \"3.0.0\"\ninfo:\n  title: Provider API\n  version: \"1.0.0\"\npaths:\n  /other:\n    get:\n      summary: Other\n      responses:\n        \"200\":\n          description: OK",
			"schema_type":         "openapi",
		}

		diffBody, _ := json.Marshal(diffReq)
		reqDiff, err := http.NewRequest("POST", p1ApiURL+"/api/v1/diff", bytes.NewReader(diffBody))
		require.NoError(t, err)
		reqDiff.Header.Set("Content-Type", "application/json")
		reqDiff.Header.Set("Authorization", "Bearer "+p1RegistryAPIToken)

		respDiff, err := http.DefaultClient.Do(reqDiff)
		require.NoError(t, err)
		defer respDiff.Body.Close()

		assert.Equal(t, http.StatusCreated, respDiff.StatusCode)
	})

	t.Run("Push non-breaking change", func(t *testing.T) {
		diffReq := map[string]interface{}{
			"diff_report": map[string]interface{}{
				"summary": map[string]interface{}{
					"breaking_count": 0,
					"safe_count":     1,
				},
				"schema_type": "openapi",
				"version":     "v1",
			},
			"is_audit_mode":       false,
			"org":                 "mcp-org",
			"provider_repo":       "mcp-org/core-repo",
			"pr_number":           3,
			"commit_sha":          "safe-sha",
			"head_schema_content": "openapi: \"3.0.0\"\ninfo:\n  title: Provider API\n  version: \"1.0.0\"\npaths:\n  /users:\n    get:\n      summary: List users\n      responses:\n        \"200\":\n          description: OK\n  /users/new:\n    get:\n      summary: New user\n      responses:\n        \"200\":\n          description: OK",
			"schema_type":         "openapi",
		}

		diffBody, _ := json.Marshal(diffReq)
		reqDiff, err := http.NewRequest("POST", p1ApiURL+"/api/v1/diff", bytes.NewReader(diffBody))
		require.NoError(t, err)
		reqDiff.Header.Set("Content-Type", "application/json")
		reqDiff.Header.Set("Authorization", "Bearer "+p1RegistryAPIToken)

		respDiff, err := http.DefaultClient.Do(reqDiff)
		require.NoError(t, err)
		defer respDiff.Body.Close()

		assert.Equal(t, http.StatusCreated, respDiff.StatusCode)
	})
}
