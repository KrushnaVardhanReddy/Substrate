// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

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
				"type":       "object",
				"properties": map[string]interface{}{},
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
		var parsedArgs struct {
			Repo  string `json:"repo"`
			Limit *int   `json:"limit,omitempty"`
		}
		if err := json.Unmarshal([]byte(args), &parsedArgs); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}
		if parsedArgs.Repo == "" {
			return "", fmt.Errorf("repo is required")
		}
		limit := 10
		if parsedArgs.Limit != nil {
			limit = *parsedArgs.Limit
		}

		registryURL := viper.GetString("REGISTRY_API_URL")
		if registryURL == "" {
			registryURL = "http://localhost:8090"
		}

		url := fmt.Sprintf("%s/api/v1/history/%s/%s?limit=%d", registryURL, req.Org, parsedArgs.Repo, limit)
		httpReq, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return "", fmt.Errorf("failed to create request: %w", err)
		}

		token := viper.GetString("REGISTRY_API_TOKEN")
		if token != "" {
			httpReq.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := http.DefaultClient.Do(httpReq)
		if err != nil {
			return "", fmt.Errorf("failed to fetch history: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return "", fmt.Errorf("registry API returned status %d: %s", resp.StatusCode, string(body))
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response: %w", err)
		}

		return string(body), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}
