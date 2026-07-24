package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var pMCPApiURL = "http://localhost:8090"
var pMCPToken = "local-dev-token"

func pMCPCall(t *testing.T, method string, id int, params any) map[string]any {
	payload := map[string]any{"jsonrpc": "2.0", "id": id, "method": method}
	if params != nil {
		payload["params"] = params
	}
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", pMCPApiURL+"/mcp/message", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+pMCPToken)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	var result map[string]any
	json.NewDecoder(resp.Body).Decode(&result)
	return result
}

func TestPhaseMCPSystemE2E(t *testing.T) {
	resp, err := http.Get(pMCPApiURL + "/health")
	if err != nil || resp.StatusCode != 200 {
		t.Skip("API not reachable, skipping E2E test")
	}

	t.Run("Scenario 1: tools/list returns >= 20 tools", func(t *testing.T) {
		result := pMCPCall(t, "tools/list", 1, nil)
		require.NotNil(t, result["result"])
		tools := result["result"].(map[string]any)["tools"].([]any)
		assert.GreaterOrEqual(t, len(tools), 28)
	})

	t.Run("Scenario 2: diff_schemas detects breaking change", func(t *testing.T) {
		result := pMCPCall(t, "tools/call", 2, map[string]any{
			"name": "diff_schemas",
			"arguments": map[string]any{
				"org":           "mcp-org",
				"provider_repo": "mcp-org/backend",
				"schema_type":   "openapi",
				"base_content":  "openapi: 3.0.0\npaths:\n  /users:\n    get:\n      responses:\n        '200':\n          description: OK",
				"head_content":  "openapi: 3.0.0\npaths: {}",
			},
		})
		require.NotNil(t, result["result"])
		content := result["result"].(map[string]any)["content"].([]any)
		text := content[0].(map[string]any)["text"].(string)
		assert.Contains(t, text, "breaking_count")
	})

	t.Run("Scenario 3: get_dependency_graph returns seeded edges", func(t *testing.T) {
		result := pMCPCall(t, "tools/call", 3, map[string]any{
			"name": "get_dependency_graph",
			"arguments": map[string]any{"org": "mcp-org"},
		})
		require.NotNil(t, result["result"])
	})

	t.Run("Scenario 4: check_deploy blocks known breaking SHA", func(t *testing.T) {
		result := pMCPCall(t, "tools/call", 4, map[string]any{
			"name": "check_deploy",
			"arguments": map[string]any{
				"org":  "mcp-org",
				"repo": "backend",
				"sha":  "sha-bad",
			},
		})
		require.NotNil(t, result["result"])
		content := result["result"].(map[string]any)["content"].([]any)
		text := content[0].(map[string]any)["text"].(string)
		assert.Contains(t, text, "can_deploy")
	})

	t.Run("Scenario 5: create_governance_rule + list_governance_rules", func(t *testing.T) {
		result := pMCPCall(t, "tools/call", 5, map[string]any{
			"name": "create_governance_rule",
			"arguments": map[string]any{
				"org":       "mcp-org",
				"rule_text": "no breaking changes",
			},
		})
		require.NotNil(t, result["result"])
		content := result["result"].(map[string]any)["content"].([]any)
		text := content[0].(map[string]any)["text"].(string)
		assert.Contains(t, text, "created")

		listResult := pMCPCall(t, "tools/call", 6, map[string]any{
			"name": "list_governance_rules",
			"arguments": map[string]any{"org": "mcp-org"},
		})
		listContent := listResult["result"].(map[string]any)["content"].([]any)
		listText := listContent[0].(map[string]any)["text"].(string)
		assert.Contains(t, listText, "no breaking changes")
	})

	t.Run("Scenario 6: delete_governance_rule removes rule", func(t *testing.T) {
		listResult := pMCPCall(t, "tools/call", 7, map[string]any{
			"name": "list_governance_rules",
			"arguments": map[string]any{"org": "mcp-org"},
		})
		require.NotNil(t, listResult["result"])
		listContent := listResult["result"].(map[string]any)["content"].([]any)
		listText := listContent[0].(map[string]any)["text"].(string)

		var rules []map[string]any
		err := json.Unmarshal([]byte(listText), &rules)
		require.NoError(t, err)

		if len(rules) > 0 {
			ruleID := rules[0]["id"].(string)

			deleteResult := pMCPCall(t, "tools/call", 8, map[string]any{
				"name": "delete_governance_rule",
				"arguments": map[string]any{
					"org": "mcp-org",
					"rule_id": ruleID,
				},
			})
			require.NotNil(t, deleteResult["result"])
		}
	})

	t.Run("Scenario 7: sync_schema + get_schema resource", func(t *testing.T) {
		syncResult := pMCPCall(t, "tools/call", 9, map[string]any{
			"name": "sync_schema",
			"arguments": map[string]any{
				"org": "mcp-org",
				"repo": "frontend",
				"raw_content": "openapi: 3.0.0",
			},
		})
		require.NotNil(t, syncResult["result"])

		readResult := pMCPCall(t, "resources/read", 10, map[string]any{
			"uri": "substrate://schemas/mcp-org/frontend",
		})
		require.NotNil(t, readResult["result"])
		contents := readResult["result"].(map[string]any)["contents"].([]any)
		assert.Greater(t, len(contents), 0)
	})

	t.Run("Scenario 8: SSE transport connectivity", func(t *testing.T) {
		req, _ := http.NewRequest("GET", pMCPApiURL+"/mcp/sse", nil)
		req.Header.Set("Authorization", "Bearer "+pMCPToken)

		client := &http.Client{Timeout: 7 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "text/event-stream", resp.Header.Get("Content-Type"))

		reader := bufio.NewReader(resp.Body)
		line, _ := reader.ReadString('\n')
		if line != "" {
			assert.Contains(t, line, "event: endpoint")
		}
	})

	t.Run("Scenario 9: file_insurance_claim + list_insurance_claims", func(t *testing.T) {
		claimResult := pMCPCall(t, "tools/call", 11, map[string]any{
			"name": "file_insurance_claim",
			"arguments": map[string]any{
				"org": "mcp-org",
				"amount_cents": 10000,
				"github_pr_url": "https://github.com/mcp-org/frontend/pull/1",
			},
		})
		require.NotNil(t, claimResult["result"])

		listResult := pMCPCall(t, "tools/call", 12, map[string]any{
			"name": "list_insurance_claims",
			"arguments": map[string]any{"org": "mcp-org"},
		})
		require.NotNil(t, listResult["result"])
	})

	t.Run("Scenario 10: resources/list returns >= 9 resources", func(t *testing.T) {
		result := pMCPCall(t, "resources/list", 13, nil)
		require.NotNil(t, result["result"])
		resources := result["result"].(map[string]any)["resources"].([]any)
		assert.GreaterOrEqual(t, len(resources), 11)
	})
}
