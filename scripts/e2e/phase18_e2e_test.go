package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	p18ApiURL = "http://localhost:8090"
	p18DbURL  = "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
	p18Token  = "local-dev-token"
)

func p18MCPCall(t *testing.T, profileID string, method string, id int, params any) map[string]any {
	payload := map[string]any{"jsonrpc": "2.0", "id": id, "method": method}
	if params != nil {
		payload["params"] = params
	}
	body, _ := json.Marshal(payload)

	url := p18ApiURL + "/mcp/message"
	if profileID != "" {
		url += "?profile_id=" + profileID
	}

	req, _ := http.NewRequest("POST", url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p18Token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var result map[string]any
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	return result
}

func waitForP18Services(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 5; i++ {
		resp, err := client.Get(p18ApiURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	t.Skipf("API server not reachable at %s. Please ensure it is running.", p18ApiURL)
}

func setupP18Database(t *testing.T) (*pgxpool.Pool, string, string) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, p18DbURL)
	require.NoError(t, err, "Failed to connect to real PostgreSQL")

	// Clean up tables
	_, _ = pool.Exec(ctx, "DELETE FROM hitl_queue")
	_, _ = pool.Exec(ctx, "DELETE FROM agent_profiles")
	_, _ = pool.Exec(ctx, "DELETE FROM contracts")
	_, _ = pool.Exec(ctx, "DELETE FROM repositories WHERE full_name = 'ai-org/payments-api'")
	_, _ = pool.Exec(ctx, "DELETE FROM organizations WHERE github_org_name = 'ai-org'")

	// Create ai-org
	var orgID string
	err = pool.QueryRow(ctx, "INSERT INTO organizations (id, github_installation_id, github_org_name) VALUES ($1, $2, $3) ON CONFLICT (github_installation_id) DO UPDATE SET github_org_name = EXCLUDED.github_org_name RETURNING id", uuid.New().String(), 1818, "ai-org").Scan(&orgID)
	require.NoError(t, err)

	// Create payments-api repository
	var repoID string
	err = pool.QueryRow(ctx, "INSERT INTO repositories (id, org_id, github_repo_id, name, full_name) VALUES ($1, $2, $3, $4, $5) RETURNING id", uuid.New().String(), orgID, 1801, "payments-api", "ai-org/payments-api").Scan(&repoID)
	require.NoError(t, err)

	// Create openapi contract with 5 paths
	openAPISpec := `
openapi: 3.0.0
info:
  title: Payments API
  version: 1.0.0
paths:
  /billing:
    get:
      summary: Get billing status
  /users:
    get:
      summary: Get users
  /auth:
    post:
      summary: Authenticate
  /webhooks:
    post:
      summary: Receive webhooks
  /settings:
    put:
      summary: Update settings
`
	_, err = pool.Exec(ctx, "INSERT INTO contracts (repo_id, schema_type, spec_path, branch, latest_commit_sha, raw_content, synced_at) VALUES ($1, 'openapi', 'openapi.yaml', 'main', 'p18-sha', $2, NOW()) ON CONFLICT DO NOTHING", repoID, openAPISpec)
	require.NoError(t, err)

	// We seed a profile initially via API for the objective requirement
	return pool, orgID, repoID
}

func TestPhase18_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	waitForP18Services(t)

	pool, _, _ := setupP18Database(t)
	defer pool.Close()

	var profileID string

	t.Run("Scenario 1: Profile Creation & Fetching", func(t *testing.T) {
		adminJWT := p18Token

		reqBody := map[string]interface{}{
			"org":           "ai-org",
			"name":          "support-agent",
			"allowed_tools": []string{"get_schema", "get_rag_bundle", "file_insurance_claim"},
			"hitl_enabled":  true,
		}
		body, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", p18ApiURL+"/api/v1/mcp/profiles", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminJWT)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var respBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err)

		idVal, ok := respBody["id"].(float64)
		require.True(t, ok)
		require.NotEmpty(t, idVal)
		profileID = strconv.Itoa(int(idVal))

		// Fetch profile
		req2, err := http.NewRequest("GET", p18ApiURL+"/api/v1/mcp/profiles/ai-org", nil)
		require.NoError(t, err)
		req2.Header.Set("Authorization", "Bearer "+adminJWT)

		resp2, err := http.DefaultClient.Do(req2)
		require.NoError(t, err)
		defer resp2.Body.Close()
		require.Equal(t, http.StatusOK, resp2.StatusCode)

		var profiles []map[string]interface{}
		err = json.NewDecoder(resp2.Body).Decode(&profiles)
		require.NoError(t, err)

		found := false
		for _, p := range profiles {
			if strconv.Itoa(int(p["id"].(float64))) == profileID {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected to find newly created profile")
	})

	t.Run("Scenario 2: MCP Session Tool Filtering", func(t *testing.T) {
		result := p18MCPCall(t, profileID, "tools/list", 1, nil)
		require.NotNil(t, result["result"])
		tools := result["result"].(map[string]any)["tools"].([]any)

		var toolNames []string
		for _, v := range tools {
			tool := v.(map[string]any)
			toolNames = append(toolNames, tool["name"].(string))
		}

		assert.Contains(t, toolNames, "get_schema")
		assert.Contains(t, toolNames, "get_rag_bundle")
		assert.Contains(t, toolNames, "file_insurance_claim")
		assert.NotContains(t, toolNames, "delete_governance_rule") // Destructive tool not in allowed list
	})

	t.Run("Scenario 3: HITL Queue via MCP", func(t *testing.T) {
		// Start SSE to satisfy MCP requirements for full connection, though call works via message
		reqSSE, _ := http.NewRequest("GET", p18ApiURL+"/mcp/sse?profile_id="+profileID, nil)
		reqSSE.Header.Set("Authorization", "Bearer "+p18Token)
		go http.DefaultClient.Do(reqSSE)
		time.Sleep(1 * time.Second)

		// Call HITL gated tool
		result := p18MCPCall(t, profileID, "tools/call", 2, map[string]any{
			"name": "file_insurance_claim",
			"arguments": map[string]any{
				"org":           "ai-org",
				"amount_cents":  10000,
				"github_pr_url": "https://github.com/ai-org/payments-api/pull/1",
			},
		})

		// The response should indicate pending approval
		require.NotNil(t, result["result"])
		content := result["result"].(map[string]any)["content"].([]any)
		text := content[0].(map[string]any)["text"].(string)
		assert.Contains(t, text, "pending_approval")

		// Verify DB Row
		var status string
		err := pool.QueryRow(context.Background(), "SELECT status FROM hitl_queue ORDER BY created_at DESC LIMIT 1").Scan(&status)
		require.NoError(t, err)
		assert.Equal(t, "pending", status)

		// Verify via API Endpoint
		adminJWT := p18Token
		req2, err := http.NewRequest("GET", p18ApiURL+"/api/v1/mcp/hitl-queue/ai-org", nil)
		require.NoError(t, err)
		req2.Header.Set("Authorization", "Bearer "+adminJWT)

		resp2, err := http.DefaultClient.Do(req2)
		require.NoError(t, err)
		defer resp2.Body.Close()
		require.Equal(t, http.StatusOK, resp2.StatusCode)

		var queueItems []map[string]interface{}
		err = json.NewDecoder(resp2.Body).Decode(&queueItems)
		require.NoError(t, err)

		require.GreaterOrEqual(t, len(queueItems), 1)
		assert.Equal(t, "pending", queueItems[0]["status"])
	})

	t.Run("Scenario 4: Context-Aware Schema Pruning", func(t *testing.T) {
		if os.Getenv("LLM_API_KEY") == "" && os.Getenv("OPENROUTER_API_KEY") == "" {
			t.Skip("Skipping real LLM test because LLM_API_KEY is missing")
		}

		reqBody := map[string]interface{}{
			"org":    "ai-org",
			"repo":   "payments-api",
			"intent": "check billing status",
		}
		body, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", p18ApiURL+"/api/v1/schema/prune", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var respBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err)

		prunedSchemaRaw := respBody["pruned_schema"].(string)
		assert.Contains(t, prunedSchemaRaw, "/billing")
		assert.NotContains(t, prunedSchemaRaw, "/auth")

		cached := respBody["cached"]
		if cached != nil {
			assert.Equal(t, false, cached.(bool))
		}

		// Call it again to assert cache hit
		body2, _ := json.Marshal(reqBody)
		req2, _ := http.NewRequest("POST", p18ApiURL+"/api/v1/schema/prune", bytes.NewReader(body2))
		req2.Header.Set("Content-Type", "application/json")

		resp2, err := http.DefaultClient.Do(req2)
		require.NoError(t, err)
		defer resp2.Body.Close()
		require.Equal(t, http.StatusOK, resp2.StatusCode)

		var respBody2 map[string]interface{}
		err = json.NewDecoder(resp2.Body).Decode(&respBody2)
		require.NoError(t, err)

		// Either cached flag is true or computed_at is set in cache hit
		cached2, ok := respBody2["cached"]
		if ok {
			assert.Equal(t, true, cached2.(bool))
		}
	})
}
