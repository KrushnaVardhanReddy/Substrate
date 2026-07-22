package services

import (
	"fmt"
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
		// Return some mock history for the org
		return fmt.Sprintf("History for %s:\n- 2023-01-01: Removed 'user_id', resulted in 3 broken builds. Remediation: added @deprecated.", req.Org), nil

	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}
