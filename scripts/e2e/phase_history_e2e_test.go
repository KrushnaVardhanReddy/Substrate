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
	pHistApiURL           = "http://localhost:8090"
	pHistDbURL            = "postgres://postgres:postgres@localhost:5432/substrate?sslmode=disable"
	pHistRegistryAPIToken = "local-dev-token"
)

func pHistWaitForServices(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 5; i++ {
		resp, err := client.Get(pHistApiURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	t.Skipf("API server not reachable at %s. Skipping E2E test.", pHistApiURL)
}

func pHistSetupDatabase(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, pHistDbURL)
	require.NoError(t, err, "Failed to connect to real PostgreSQL")

	// Clean tables
	_, err = pool.Exec(ctx, "DELETE FROM breaking_change_history")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM contracts")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM dependencies")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM repositories")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM organizations")
	require.NoError(t, err)

	return pool
}

func TestPhaseHistoryE2E(t *testing.T) {
	pHistWaitForServices(t)

	pool := pHistSetupDatabase(t)
	defer pool.Close()

	ctx := context.Background()

	// Seed DB state
	var orgID string
	err := pool.QueryRow(ctx, "INSERT INTO organizations (github_installation_id, github_org_name) VALUES ($1, $2) RETURNING id", 11111, "hist-org").Scan(&orgID)
	require.NoError(t, err)

	var backendRepoID string
	err = pool.QueryRow(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name) VALUES ($1, $2, $3, $4) RETURNING id", orgID, 22221, "backend", "hist-org/backend").Scan(&backendRepoID)
	require.NoError(t, err)

	var frontendRepoID string
	err = pool.QueryRow(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name) VALUES ($1, $2, $3, $4) RETURNING id", orgID, 22222, "frontend", "hist-org/frontend").Scan(&frontendRepoID)
	require.NoError(t, err)

	var cleanRepoID string
	err = pool.QueryRow(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name) VALUES ($1, $2, $3, $4) RETURNING id", orgID, 22223, "clean-repo", "hist-org/clean-repo").Scan(&cleanRepoID)
	require.NoError(t, err)

	t.Run("Scenario 1: Store and Retrieve Breaking Change History", func(t *testing.T) {
		payload := map[string]interface{}{
			"repo_id":   backendRepoID,
			"org_name":  "hist-org",
			"repo_name": "backend", // Handler uses org, repo as path params on GET, but POST needs them as fields
			"git_sha":   "sha-hist-test",
			"breaking_changes": []map[string]interface{}{
				{"description": "Removed field id from user response"},
			},
		}

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", pHistApiURL+"/api/v1/history", bytes.NewReader(body))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+pHistRegistryAPIToken)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		// GET /api/v1/history/hist-org/backend
		req, err = http.NewRequest("GET", pHistApiURL+"/api/v1/history/hist-org/backend", nil)
		require.NoError(t, err)

		req.Header.Set("Authorization", "Bearer "+pHistRegistryAPIToken)
		resp, err = client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(result), 1)

		assert.Equal(t, "sha-hist-test", result[0]["git_sha"])

		bc, ok := result[0]["breaking_changes"].([]interface{})
		require.True(t, ok)
		assert.NotEmpty(t, bc)

		// Add second entry for next scenario
		payload2 := map[string]interface{}{
			"repo_id":   frontendRepoID,
			"org_name":  "hist-org",
			"repo_name": "frontend",
			"git_sha":   "sha-hist-test-2",
			"breaking_changes": []map[string]interface{}{
				{"description": "Changed auth method"},
			},
		}

		body2, err := json.Marshal(payload2)
		require.NoError(t, err)
		req2, err := http.NewRequest("POST", pHistApiURL+"/api/v1/history", bytes.NewReader(body2))
		require.NoError(t, err)
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer "+pHistRegistryAPIToken)
		resp2, err := client.Do(req2)
		require.NoError(t, err)
		defer resp2.Body.Close()
		require.Equal(t, http.StatusCreated, resp2.StatusCode)
	})

	t.Run("Scenario 2: Changes Feed Returns Recent Events", func(t *testing.T) {
		// Calculate a 'since' timestamp to be 1 hour ago
		sinceStr := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
		req, err := http.NewRequest("GET", pHistApiURL+"/api/v1/changes?since="+sinceStr, nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+pHistRegistryAPIToken)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		found := false
		for _, event := range result {
			if event["repo_name"] == "backend" {
				found = true
				break
			}
		}
		assert.True(t, found, "Response should contain at least the repo name 'backend'")
	})

	t.Run("Scenario 3: History Returns Empty for Repo With No History", func(t *testing.T) {
		req, err := http.NewRequest("GET", pHistApiURL+"/api/v1/history/hist-org/clean-repo", nil)
		require.NoError(t, err)

		req.Header.Set("Authorization", "Bearer "+pHistRegistryAPIToken)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}
