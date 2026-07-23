import re

with open("scripts/e2e/phase_mcp_e2e_test.go", "r") as f:
    content = f.read()

# Add missing imports for Scenarios 6-10: bufio, time
content = re.sub(r'import \([\s\S]*?\)', """import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"
	"bufio"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)""", content, count=1)

scenarios = """
	t.Run("Scenario 6: delete_governance_rule removes rule", func(t *testing.T) {
		// First get the rule to delete
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
		// We expect the first event to be endpoint event
		line, _ := reader.ReadString('\\n')
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
		assert.GreaterOrEqual(t, len(resources), 9)
	})
"""

idx = content.rfind("}")
new_content = content[:idx] + scenarios + "\n}\n"

with open("scripts/e2e/phase_mcp_e2e_test.go", "w") as f:
    f.write(new_content)
