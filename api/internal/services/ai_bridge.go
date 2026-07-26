package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/viper"
)

// MCP Tool definitions for OpenAI
var mcpTools = []map[string]interface{}{
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "analyze_breaking_changes",
			"description": "Analyzes the current and proposed schemas to find breaking changes.",
			"parameters": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	},
	{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "get_breaking_change_history",
			"description": "Gets past breaking changes and their resolutions for this repository.",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"repo": map[string]interface{}{
						"type":        "string",
						"description": "The repository to check (e.g. 'myorg/backend-api')",
					},
					"limit": map[string]interface{}{
						"type":        "integer",
						"description": "Maximum number of records to return (default 10)",
					},
				},
				"required": []string{"repo"},
			},
		},
	},
}

type chatMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	Name       string     `json:"name,omitempty"`
	ToolCalls  []toolCall `json:"tool_calls,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
}

type toolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type chatRequest struct {
	Model    string                   `json:"model"`
	Stream   bool                     `json:"stream"`
	Messages []chatMessage            `json:"messages"`
	Tools    []map[string]interface{} `json:"tools,omitempty"`
}

func executeTool(name string, args string, req AIAnalyzeRequest) (string, error) {
	switch name {
	case "analyze_breaking_changes":
		if req.CurrentSchema != req.ProposedSchema {
			return "[{\"rule_id\": \"mock-break\", \"description\": \"Field removed or modified\", \"severity\": \"BREAKING\"}]", nil
		}
		return "[]", nil

	case "get_breaking_change_history":
		var params struct {
			Repo  string `json:"repo"`
			Limit int    `json:"limit"`
		}
		if err := json.Unmarshal([]byte(args), &params); err != nil {
			return "", err
		}
		if params.Repo == "" {
			return "", fmt.Errorf("repo is required")
		}
		parts := strings.SplitN(params.Repo, "/", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("repo must be in format org/repo")
		}

		limit := params.Limit
		if limit == 0 {
			limit = 10
		}

		registryURL := viper.GetString("REGISTRY_API_URL")
		if registryURL == "" {
			registryURL = "http://localhost:8090"
		}

		url := fmt.Sprintf("%s/api/v1/history/%s/%s?limit=%d", registryURL, parts[0], parts[1], limit)

		httpReq, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return "", err
		}

		token := viper.GetString("REGISTRY_API_TOKEN")
		if token != "" {
			httpReq.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		return string(body), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}
