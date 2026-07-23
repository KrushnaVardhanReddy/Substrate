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
	pRoiApiURL           = "http://localhost:8090"
	pRoiDbURL            = "postgres://postgres:postgres@localhost:5432/substrate?sslmode=disable"
	pRoiRegistryAPIToken = "local-dev-token"
)

func pRoiWaitForServices(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 5; i++ {
		resp, err := client.Get(pRoiApiURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	t.Skipf("API server not reachable at %s. Skipping E2E test.", pRoiApiURL)
}

func pRoiSetupDatabase(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, pRoiDbURL)
	require.NoError(t, err, "Failed to connect to real PostgreSQL")

	// Clean tables
	_, err = pool.Exec(ctx, "DELETE FROM diff_reports")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM drift_anomalies")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM organizations")
	require.NoError(t, err)

	return pool
}

func TestPhaseRoiE2E(t *testing.T) {
	pRoiWaitForServices(t)

	pool := pRoiSetupDatabase(t)
	defer pool.Close()

	ctx := context.Background()

	// Seed DB state
	var orgID string
	err := pool.QueryRow(ctx, "INSERT INTO organizations (github_installation_id, github_org_name) VALUES ($1, $2) RETURNING id", 12345, "roi-org").Scan(&orgID)
	require.NoError(t, err)

	// Seed diff_reports
	// Total prevented outages = 2 (is_audit_mode=true or status='blocked', created_at within 30 days)
	_, err = pool.Exec(ctx, `
		INSERT INTO diff_reports (id, org_name, repo_name, is_audit_mode, report_data, created_at)
		VALUES
			(gen_random_uuid(), 'roi-org', 'repo1', false, '{"status": "blocked"}', NOW() - INTERVAL '1 day'),
			(gen_random_uuid(), 'roi-org', 'repo2', true, '{"status": "passed"}', NOW() - INTERVAL '2 days'),
			(gen_random_uuid(), 'roi-org', 'repo3', false, '{"status": "passed"}', NOW() - INTERVAL '3 days')
	`)
	require.NoError(t, err)

	// Seed drift_anomalies
	// Total undocumented endpoints = 3
	_, err = pool.Exec(ctx, `
		INSERT INTO drift_anomalies (id, org_name, repo_name, method, path, error_message, timestamp)
		VALUES
			(gen_random_uuid(), 'roi-org', 'repo1', 'GET', '/api/1', 'msg', NOW() - INTERVAL '1 day'),
			(gen_random_uuid(), 'roi-org', 'repo1', 'POST', '/api/2', 'msg', NOW() - INTERVAL '2 days'),
			(gen_random_uuid(), 'roi-org', 'repo2', 'PUT', '/api/3', 'msg', NOW() - INTERVAL '3 days')
	`)
	require.NoError(t, err)

	t.Run("Scenario 1: ROI Fetch", func(t *testing.T) {
		req, err := http.NewRequest("GET", pRoiApiURL+"/api/v1/telemetry/roi/roi-org", nil)
		require.NoError(t, err)

		req.Header.Set("Authorization", "Bearer "+pRoiRegistryAPIToken)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.Equal(t, float64(2), result["total_prevented_outages"])
		assert.Equal(t, float64(3), result["total_undocumented_endpoints"])

		// Expected HoursSaved = 2 * 4 = 8
		assert.Equal(t, float64(8), result["hours_saved"])

		// Expected EstimatedDollarValueSaved = 8 * 100 = 800
		assert.Equal(t, float64(800), result["estimated_dollar_value_saved"])
	})

	t.Run("Scenario 2: Predict Cost", func(t *testing.T) {
		payload := map[string]interface{}{
			"base_schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{"type": "integer"},
				},
			},
			"proposed_schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id":          map[string]interface{}{"type": "integer"},
					"name":        map[string]interface{}{"type": "string"},
					"description": map[string]interface{}{"type": "string"},
				},
			},
			"endpoint_path": "/api/v1/test",
		}

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", pRoiApiURL+"/api/v1/finops/predict", bytes.NewReader(body))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+pRoiRegistryAPIToken)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Base schema byte size: id(integer) = 8 -> 8
		// Proposed schema byte size: id(8) + name(string:50) + description(string:50) -> 108

		assert.Equal(t, float64(8), result["base_bytes"])
		assert.Equal(t, float64(108), result["proposed_bytes"])
		assert.Greater(t, result["monthly_cost_diff"].(float64), float64(0))
		assert.GreaterOrEqual(t, result["rps_used"].(float64), float64(0))
	})
}
