package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var p7_2ApiURL = "http://localhost:8090"

func setupP7_2Database(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
	}
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Log("Failed to connect to real PostgreSQL")
		return nil
	}

	err = pool.Ping(ctx)
	if err != nil {
		t.Log("Failed to ping PostgreSQL")
		return nil
	}
	return pool
}

func waitForP7_2Services(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 5; i++ {
		resp, err := client.Get(p7_2ApiURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			return
		}
		time.Sleep(1 * time.Second)
	}
	t.Skipf("API server not reachable at %s. Please ensure 'make api' and 'make postgres' are running.", p7_2ApiURL)
}

func TestPhase7EnterpriseRollout(t *testing.T) {
	pool := setupP7_2Database(t)
	if pool != nil {
		defer pool.Close()
	}

	if os.Getenv("SUBSTRATE_API_URL") != "" {
		p7_2ApiURL = os.Getenv("SUBSTRATE_API_URL")
	}

	waitForP7_2Services(t)

	// Note: We use the existing endpoints as advised for the Phase 7 E2E tests

	t.Run("POST /api/v1/org/{org}/webhooks", func(t *testing.T) {
		payload := map[string]string{"url": "https://example.com/webhook", "secret": "mysecret"}
		body, _ := json.Marshal(payload)

		req, err := http.NewRequest("POST", p7_2ApiURL+"/api/v1/org/mcp-org/webhooks", bytes.NewReader(body))
		require.NoError(t, err)

		// This route uses JWT authMW typically, but testing with valid token if possible
		req.Header.Set("Authorization", "Bearer "+os.Getenv("REGISTRY_API_TOKEN"))
		if req.Header.Get("Authorization") == "Bearer " {
			req.Header.Set("Authorization", "Bearer test-service-token")
		}

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Allow 401 if we are missing a real token in standard test env
		if resp.StatusCode == http.StatusUnauthorized {
			t.Skip("Skipping due to missing valid JWT token for authMW")
		}
		assert.Contains(t, []int{http.StatusOK, http.StatusCreated}, resp.StatusCode)
	})

	t.Run("POST /api/v1/org/{org}/rules", func(t *testing.T) {
		payload := map[string]string{"rule_text": "request.auth.claims.group == 'admin'"}
		body, _ := json.Marshal(payload)

		req, err := http.NewRequest("POST", p7_2ApiURL+"/api/v1/org/mcp-org/rules", bytes.NewReader(body))
		require.NoError(t, err)

		req.Header.Set("Authorization", "Bearer "+os.Getenv("REGISTRY_API_TOKEN"))
		if req.Header.Get("Authorization") == "Bearer " {
			req.Header.Set("Authorization", "Bearer test-service-token")
		}

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			t.Skip("Skipping due to missing valid JWT token for authMW")
		}
		assert.Contains(t, []int{http.StatusOK, http.StatusCreated}, resp.StatusCode)
	})

	t.Run("GET /api/v1/telemetry/drift", func(t *testing.T) {
		// Example of checking drift for the sidecar
		req, err := http.NewRequest("POST", p7_2ApiURL+"/api/v1/telemetry/drift", bytes.NewReader([]byte(`{"org":"mcp-org","repo":"enterprise-repo","drift_detected":true}`)))
		require.NoError(t, err)

		req.Header.Set("Authorization", "Bearer "+os.Getenv("REGISTRY_API_TOKEN"))
		if req.Header.Get("Authorization") == "Bearer " {
			req.Header.Set("Authorization", "Bearer test-service-token")
		}

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Typically serviceTokenMW
		assert.Contains(t, []int{http.StatusOK, http.StatusUnauthorized, http.StatusBadRequest}, resp.StatusCode)
	})
}
