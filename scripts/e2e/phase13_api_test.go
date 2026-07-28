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
	p13RegistryAPIToken = "local-dev-token"
	p13ApiURL           = "http://localhost:8090"
	p13DbURL            = "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
)

func waitForP13Services(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 5; i++ {
		resp, err := client.Get(p13ApiURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	t.Skipf("API server not reachable at %s. Skipping test in restricted environment.", p13ApiURL)
}

func setupP13Database(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, p13DbURL)
	require.NoError(t, err, "Failed to connect to real PostgreSQL")

	// Clean up and seed data for MCP org
	_, err = pool.Exec(ctx, "DELETE FROM organizations WHERE github_org_name = 'mcp-org'")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO organizations (id, github_installation_id, github_org_name)
		VALUES ('11111111-1111-1111-1111-111111111111', 123456, 'mcp-org')
		ON CONFLICT (id) DO NOTHING;
	`)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO repositories (id, org_id, github_repo_id, name, full_name)
		VALUES ('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 101, 'core-repo', 'mcp-org/core-repo')
		ON CONFLICT (id) DO NOTHING;
	`)
	require.NoError(t, err)

	return pool
}

func TestPhase13_API_E2E(t *testing.T) {
	waitForP13Services(t)

	pool := setupP13Database(t)
	defer pool.Close()

	t.Run("TestGodMode_FinOpsPredict", func(t *testing.T) {
		baseSchema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id":    map[string]interface{}{"type": "string"},
				"email": map[string]interface{}{"type": "string"},
			},
		}
		proposedSchema := map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"id": map[string]interface{}{"type": "string"},
				"email": map[string]interface{}{"type": "string"},
				"newField": map[string]interface{}{"type": "string"},
			},
		}

		payload := map[string]interface{}{
			"base_schema":     baseSchema,
			"proposed_schema": proposedSchema,
			"endpoint_path":   "/users",
		}
		bodyBytes, _ := json.Marshal(payload)

		req, err := http.NewRequest("POST", p13ApiURL+"/api/v1/finops/predict", bytes.NewReader(bodyBytes))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+p13RegistryAPIToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var respData map[string]interface{}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&respData))

		costDiff, ok := respData["monthly_cost_diff"].(float64)
		require.True(t, ok)
		assert.Greater(t, costDiff, 0.0, "Cost diff should be positive because we added a field")

		rpsUsed, ok := respData["rps_used"].(float64)
		require.True(t, ok)
		assert.Greater(t, rpsUsed, 0.0, "RPS used should be greater than 0")
	})
}
