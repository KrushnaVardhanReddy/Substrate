package mcp

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestHandleMessage_Initialize(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Initialize should return correct protocol version and capabilities",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()
			req := `{"jsonrpc": "2.0", "id": 1, "method": "initialize"}`
			respBytes := server.HandleMessage([]byte(req))

			var resp JSONRPCResponse
			if err := json.Unmarshal(respBytes, &resp); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if resp.Error != nil {
				t.Fatalf("Expected no error, got: %v", resp.Error)
			}

			if *resp.ID != 1 {
				t.Errorf("Expected id 1, got %v", *resp.ID)
			}

			var result map[string]any
			if err := json.Unmarshal(resp.Result, &result); err != nil {
				t.Fatalf("Failed to parse result: %v", err)
			}

			if result["protocolVersion"] != "2024-11-05" {
				t.Errorf("Unexpected protocolVersion: %v", result["protocolVersion"])
			}
		})
	}
}

func TestHandleMessage_ToolsList(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Tools list should return registered tools",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()
			server.RegisterTool(Tool{
				Name:        "test_tool",
				Description: "A test tool",
				InputSchema: map[string]any{"type": "object"},
				Handler: func(params json.RawMessage) (any, error) {
					return "success", nil
				},
			})

			req := `{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}`
			respBytes := server.HandleMessage([]byte(req))

			var resp JSONRPCResponse
			if err := json.Unmarshal(respBytes, &resp); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			var result struct {
				Tools []struct {
					Name        string         `json:"name"`
					Description string         `json:"description"`
					InputSchema map[string]any `json:"inputSchema"`
				} `json:"tools"`
			}
			if err := json.Unmarshal(resp.Result, &result); err != nil {
				t.Fatalf("Failed to parse result: %v", err)
			}

			if len(result.Tools) != 1 {
				t.Fatalf("Expected 1 tool, got %d", len(result.Tools))
			}
			if result.Tools[0].Name != "test_tool" {
				t.Errorf("Expected tool name 'test_tool', got %s", result.Tools[0].Name)
			}
		})
	}
}

func TestHandleMessage_ToolsCall(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Tools call should invoke handler correctly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()
			server.RegisterTool(Tool{
				Name:        "test_tool",
				Description: "A test tool",
				InputSchema: map[string]any{"type": "object"},
				Handler: func(params json.RawMessage) (any, error) {
					return "Hello from tool", nil
				},
			})

			req := `{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "test_tool", "arguments": {}}}`
			respBytes := server.HandleMessage([]byte(req))

			var resp JSONRPCResponse
			if err := json.Unmarshal(respBytes, &resp); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if resp.Error != nil {
				t.Fatalf("Expected no error, got: %v", resp.Error)
			}

			resultStr := string(resp.Result)
			if !strings.Contains(resultStr, "Hello from tool") {
				t.Errorf("Expected 'Hello from tool' in result, got: %s", resultStr)
			}
		})
	}
}

func TestHandleMessage_InvalidMethod(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "Invalid method should return -32601",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer()
			req := `{"jsonrpc": "2.0", "id": 4, "method": "unknown_method"}`
			respBytes := server.HandleMessage([]byte(req))

			var resp JSONRPCResponse
			if err := json.Unmarshal(respBytes, &resp); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			if resp.Error == nil {
				t.Fatalf("Expected error for unknown method")
			}
			if resp.Error.Code != -32601 {
				t.Errorf("Expected error code -32601, got %d", resp.Error.Code)
			}
		})
	}
}
