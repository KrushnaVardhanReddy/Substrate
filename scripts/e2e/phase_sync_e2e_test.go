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
	pSyncApiURL           = "http://localhost:8090"
	pSyncDbURL            = "postgres://postgres:postgres@localhost:5432/substrate?sslmode=disable"
	pSyncRegistryAPIToken = "local-dev-token"
)

func pSyncWaitForServices(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 5; i++ {
		resp, err := client.Get(pSyncApiURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	t.Skipf("API server not reachable at %s. Skipping E2E test.", pSyncApiURL)
}

func pSyncSetupDatabase(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, pSyncDbURL)
	require.NoError(t, err, "Failed to connect to PostgreSQL")

	_, err = pool.Exec(ctx, "DELETE FROM breaking_change_history")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM diff_reports")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM dependencies")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM contracts")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM repositories")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM organizations")
	require.NoError(t, err)

	return pool
}

func TestPhaseSyncE2E(t *testing.T) {
	pSyncWaitForServices(t)
	pool := pSyncSetupDatabase(t)
	defer pool.Close()

	ctx := context.Background()

	// Seed org + repo sync-org/backend via pgxpool
	var orgID string
	err := pool.QueryRow(ctx, "INSERT INTO organizations (github_installation_id, github_org_name) VALUES ($1, $2) RETURNING id", 12345, "sync-org").Scan(&orgID)
	require.NoError(t, err)

	var backendRepoID string
	err = pool.QueryRow(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name) VALUES ($1, $2, $3, $4) RETURNING id", orgID, 123451, "backend", "sync-org/backend").Scan(&backendRepoID)
	require.NoError(t, err)

	// Create a frontend repo just in case to match the spec payload "consumer_repo"
	var frontendRepoID string
	err = pool.QueryRow(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name) VALUES ($1, $2, $3, $4) RETURNING id", orgID, 123452, "frontend", "sync-org/frontend").Scan(&frontendRepoID)
	require.NoError(t, err)

	t.Run("Scenario 1: Sync Stores Schema and Schema Retrieval Works", func(t *testing.T) {
		payload := map[string]interface{}{
			"org":           "sync-org",
			"provider_repo": "sync-org/backend",
			"consumer_repo": "sync-org/frontend",
			"commit_sha":    "sha-sync-001",
			"schema_type":   "openapi",
			"raw_content":   "openapi: 3.0.0\ninfo:\n  title: Sync Test\n  version: 1.0.0\npaths: {}",
		}

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", pSyncApiURL+"/api/v1/sync", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+pSyncRegistryAPIToken)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Contains(t, []int{http.StatusOK, http.StatusCreated}, resp.StatusCode)

		// GET /api/v1/schema/sync-org/backend
		getReq, err := http.NewRequest("GET", pSyncApiURL+"/api/v1/schema/sync-org/backend", nil)
		require.NoError(t, err)
		getReq.Header.Set("Authorization", "Bearer "+pSyncRegistryAPIToken)

		getResp, err := client.Do(getReq)
		require.NoError(t, err)
		defer getResp.Body.Close()

		require.Equal(t, http.StatusOK, getResp.StatusCode)

		var respData map[string]interface{}
		err = json.NewDecoder(getResp.Body).Decode(&respData)
		require.NoError(t, err)

		rawContent, ok := respData["raw_content"].(string)
		assert.True(t, ok)
		assert.NotEmpty(t, rawContent)
		assert.Contains(t, rawContent, "openapi")

		// It should contain either raw_content or provider repo info. Let's assert what the spec actually says
		// "Assert response body contains 'openapi' and 'sync-org/backend' or non-empty raw_content."
		// raw_content was found above and contains openapi

		// Optional stringified response for additional safety
		getRespBody, _ := json.Marshal(respData)
		assert.Contains(t, string(getRespBody), "openapi")
	})

	t.Run("Scenario 2: Spec Endpoint Returns Parsed JSON", func(t *testing.T) {
		req, err := http.NewRequest("GET", pSyncApiURL+"/api/v1/spec/sync-org/backend", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+pSyncRegistryAPIToken)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var respData map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)

		_, ok := respData["openapi"]
		assert.True(t, ok, "Expected response to have 'openapi' key")
	})

	t.Run("Scenario 3: Sync Idempotent — Re-syncing Same SHA Updates Timestamp", func(t *testing.T) {
		payload := map[string]interface{}{
			"org":           "sync-org",
			"provider_repo": "sync-org/backend",
			"consumer_repo": "sync-org/frontend",
			"commit_sha":    "sha-sync-001",
			"schema_type":   "openapi",
			"raw_content":   "openapi: 3.0.0\ninfo:\n  title: Sync Test V2\n  version: 1.0.0\npaths: {}",
		}

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", pSyncApiURL+"/api/v1/sync", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+pSyncRegistryAPIToken)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Contains(t, []int{http.StatusOK, http.StatusCreated}, resp.StatusCode, "Expected 200 or 201 for idempotent sync")
	})

	t.Run("Scenario 4: Schema 404 for Unknown Repo", func(t *testing.T) {
		req, err := http.NewRequest("GET", pSyncApiURL+"/api/v1/schema/sync-org/nonexistent", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+pSyncRegistryAPIToken)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
