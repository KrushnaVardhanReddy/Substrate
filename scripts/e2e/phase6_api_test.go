package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhase6QAShadowAPI(t *testing.T) {
	// Skip if services are not available
	_, err := http.Get(apiURL + "/health")
	if err != nil {
		t.Skip("API not reachable")
	}

	// Wait for services to be ready
	waitForServices(t)

	// Test GET /api/v1/qa/postman/{org}/{repo}
	t.Run("Postman Collection Generation", func(t *testing.T) {
		req, _ := http.NewRequest("GET", apiURL + "/api/v1/qa/postman/mcp-org/shadow-api-repo", nil)
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

		req, _ := http.NewRequest("POST", apiURL + "/api/v1/qa/shadow/replay", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test GET /api/v1/qa/coverage/{org}/{repo}
	t.Run("Shadow API Coverage", func(t *testing.T) {
		req, _ := http.NewRequest("GET", apiURL + "/api/v1/qa/coverage/mcp-org/shadow-api-repo", nil)

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
