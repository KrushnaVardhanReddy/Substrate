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
	p6DbURL = "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
)

func setupP6Database(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, p6DbURL)
	if err != nil {
		t.Skipf("Failed to connect to PostgreSQL: %v", err)
	}
	// Ensure mcp-org exists so QAPostmanHandler can find it
	_, err = pool.Exec(ctx, `
		INSERT INTO organizations (github_installation_id, github_org_name)
		VALUES (606, 'mcp-org')
		ON CONFLICT (github_installation_id) DO UPDATE SET github_org_name = EXCLUDED.github_org_name
	`)
	require.NoError(t, err)
	return pool
}

func TestPhase6QAShadowAPI(t *testing.T) {
	// Skip if services are not available
	_, err := http.Get(apiURL + "/health")
	if err != nil {
		t.Skip("API not reachable")
	}

	// Wait for services to be ready
	waitForServices(t)

	// Seed org so handler can find it
	pool := setupP6Database(t)
	defer pool.Close()

	// Test GET /api/v1/qa/postman/{org}/{repo}
	t.Run("Postman Collection Generation", func(t *testing.T) {
		req, _ := http.NewRequest("GET", apiURL+"/api/v1/qa/postman/mcp-org/shadow-api-repo", nil)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		require.NoError(t, err)

		info, ok := data["info"].(map[string]interface{})
		require.True(t, ok)
		assert.Contains(t, info["name"], "Auto-Generated Collection")
	})

	// Test POST /api/v1/qa/shadow/replay
	t.Run("Shadow API Replay", func(t *testing.T) {
		payload := map[string]string{
			"org":       "mcp-org",
			"repo":      "shadow-api-repo",
			"timestamp": "2026-07-01",
		}
		body, _ := json.Marshal(payload)

		req, _ := http.NewRequest("POST", apiURL+"/api/v1/qa/shadow/replay", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test GET /api/v1/qa/coverage/{org}/{repo}
	t.Run("Shadow API Coverage", func(t *testing.T) {
		req, _ := http.NewRequest("GET", apiURL+"/api/v1/qa/coverage/mcp-org/shadow-api-repo", nil)

		// Add small delay to ensure replay is processed
		time.Sleep(500 * time.Millisecond)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		require.NoError(t, err)

		assert.NotNil(t, data["score"])
		assert.NotNil(t, data["details"])
	})
}
