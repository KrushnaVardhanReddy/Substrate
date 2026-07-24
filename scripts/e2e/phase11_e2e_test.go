package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	p11ApiURL           = "http://localhost:8090"
	p11DbURL            = "postgres://postgres:postgres@localhost:5432/substrate?sslmode=disable"
	p11RegistryAPIToken = "local-dev-token"
)

func waitForP11Services(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 5; i++ {
		resp, err := client.Get(p11ApiURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	t.Skipf("API server not reachable at %s. Please ensure 'make api' and 'make postgres' are running.", p11ApiURL)
}

func setupP11Database(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, p11DbURL)
	if err != nil {
		t.Skip("Failed to connect to real PostgreSQL")
	}

	// Clean up tables relevant to Phase 11
	_, err = pool.Exec(ctx, "DELETE FROM drift_anomalies")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM endpoint_traffic")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM governance_rules")
	require.NoError(t, err)
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

func TestPhase11SystemE2E(t *testing.T) {
	waitForP11Services(t)

	pool := setupP11Database(t)
	defer pool.Close()

	ctx := context.Background()
	var orgID string

	// Scenario 1: Dependency Graph Returns Correct Edges
	t.Run("Scenario 1: Dependency Graph returns correct edges", func(t *testing.T) {
		// 1. Insert Org
		err := pool.QueryRow(ctx, "INSERT INTO organizations (github_installation_id, github_org_name) VALUES ($1, $2) RETURNING id", 11001, "p11-org").Scan(&orgID)
		require.NoError(t, err)

		// 2. Insert Provider Repo
		var providerRepoID string
		err = pool.QueryRow(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name) VALUES ($1, $2, $3, $4) RETURNING id", orgID, 11002, "backend", "p11-org/backend").Scan(&providerRepoID)
		require.NoError(t, err)

		// 3. Insert Consumer Repo
		var consumerRepoID string
		err = pool.QueryRow(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name) VALUES ($1, $2, $3, $4) RETURNING id", orgID, 11003, "frontend", "p11-org/frontend").Scan(&consumerRepoID)
		require.NoError(t, err)

		// 4. Insert Contract for Provider
		var contractID string
		err = pool.QueryRow(ctx, "INSERT INTO contracts (repo_id, schema_type, spec_path, raw_content) VALUES ($1, $2, $3, $4) RETURNING id", providerRepoID, "openapi", "openapi.yaml", "mock-content").Scan(&contractID)
		require.NoError(t, err)

		// 5. Map Consumer to Contract
		_, err = pool.Exec(ctx, "INSERT INTO dependencies (consumer_repo_id, provider_contract_id, status) VALUES ($1, $2, $3)", consumerRepoID, contractID, "active")
		require.NoError(t, err)

		// Make HTTP request
		req, err := http.NewRequest("GET", p11ApiURL+"/api/v1/graph/p11-org", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+p11RegistryAPIToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)
		require.Greater(t, len(result), 0)

		assert.Equal(t, "p11-org/frontend", result[0]["consumer"])
		assert.Equal(t, "p11-org/backend", result[0]["provider"])
	})

	// Scenario 2: Impact Analysis Enumerates Downstream Consumers
	t.Run("Scenario 2: Impact analysis enumerates downstream consumers", func(t *testing.T) {
		req, err := http.NewRequest("GET", p11ApiURL+"/api/v1/impact/p11-org/backend", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+p11RegistryAPIToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Contains(t, fmt.Sprint(result["impacted_repos"]), "p11-org/frontend")
	})

	// Scenario 3: SSE Stream Connects and Delivers Heartbeat
	t.Run("Scenario 3: SSE stream connects and delivers heartbeat", func(t *testing.T) {
		client := &http.Client{}

		reqCtx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
		defer cancel()

		req, err := http.NewRequestWithContext(reqCtx, "GET", p11ApiURL+"/api/v1/events", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+p11RegistryAPIToken)

		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

		scanner := bufio.NewScanner(resp.Body)
		receivedHeartbeat := false

		// We will loop to read lines
		// Context cancellation will close the response body and scanner.Scan() will return false
		for scanner.Scan() {
			line := scanner.Text()
			if len(line) > 0 && (line == ": heartbeat" || line == "data: {\"type\":\"heartbeat\"}" || line == "data: heartbeat" || (len(line) > 5 && line[:5] == "data:")) {
				receivedHeartbeat = true
				cancel()
				break
			}
		}

		assert.True(t, receivedHeartbeat, "Should receive at least one data line (heartbeat) from SSE stream")

		err = resp.Body.Close()
		require.NoError(t, err, "Connection should close cleanly")
	})

	// Scenario 4: Diff Report Persistence and Retrieval
	t.Run("Scenario 4: Diff report persistence and retrieval", func(t *testing.T) {
		// Create Diff
		payload := []byte(`{"diff_report": {"summary": {"breaking_count": 1}}, "org": "p11-org"}`)
		req, err := http.NewRequest("POST", p11ApiURL+"/api/v1/diff", bytes.NewReader(payload))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+p11RegistryAPIToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var createResult map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&createResult)
		require.NoError(t, err)

		diffID, ok := createResult["id"].(string)

		require.NotEmpty(t, diffID)
		_ = ok

		// Get Diff
		reqGet, err := http.NewRequest("GET", p11ApiURL+"/api/v1/diff/"+diffID, nil)
		require.NoError(t, err)
		reqGet.Header.Set("Authorization", "Bearer "+p11RegistryAPIToken)

		respGet, err := http.DefaultClient.Do(reqGet)
		require.NoError(t, err)
		defer respGet.Body.Close()

		assert.Equal(t, http.StatusOK, respGet.StatusCode)
		var reportData map[string]interface{}
		err = json.NewDecoder(respGet.Body).Decode(&reportData)
		require.NoError(t, err)

		require.NotEmpty(t, reportData)

	})
}
