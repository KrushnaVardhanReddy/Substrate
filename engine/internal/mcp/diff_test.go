package mcp

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestDiffMCPState(t *testing.T) {
	tests := []struct {
		name    string
		base    *MCPState
		current *MCPState
		want    []BreakingChange
	}{
		{
			name: "no changes",
			base: &MCPState{
				Tools: map[string]ToolDef{
					"tool1": {
						Name: "tool1",
						InputSchema: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"param1": map[string]any{"type": "string"},
							},
						},
					},
				},
			},
			current: &MCPState{
				Tools: map[string]ToolDef{
					"tool1": {
						Name: "tool1",
						InputSchema: map[string]any{
							"type": "object",
							"properties": map[string]any{
								"param1": map[string]any{"type": "string"},
							},
						},
					},
				},
			},
			want: []BreakingChange{},
		},
		{
			name: "tool removed",
			base: &MCPState{
				Tools: map[string]ToolDef{
					"tool1": {Name: "tool1"},
				},
			},
			current: &MCPState{
				Tools: map[string]ToolDef{},
			},
			want: []BreakingChange{
				{Type: "TOOL_REMOVED", Severity: SeverityHigh},
			},
		},
		{
			name: "parameter removed",
			base: &MCPState{
				Tools: map[string]ToolDef{
					"tool1": {
						Name: "tool1",
						InputSchema: map[string]any{
							"properties": map[string]any{
								"param1": map[string]any{"type": "string"},
							},
						},
					},
				},
			},
			current: &MCPState{
				Tools: map[string]ToolDef{
					"tool1": {
						Name: "tool1",
						InputSchema: map[string]any{
							"properties": map[string]any{},
						},
					},
				},
			},
			want: []BreakingChange{
				{Type: "PARAMETER_REMOVED", Severity: SeverityHigh},
			},
		},
		{
			name: "parameter type changed",
			base: &MCPState{
				Tools: map[string]ToolDef{
					"tool1": {
						Name: "tool1",
						InputSchema: map[string]any{
							"properties": map[string]any{
								"param1": map[string]any{"type": "string"},
							},
						},
					},
				},
			},
			current: &MCPState{
				Tools: map[string]ToolDef{
					"tool1": {
						Name: "tool1",
						InputSchema: map[string]any{
							"properties": map[string]any{
								"param1": map[string]any{"type": "integer"},
							},
						},
					},
				},
			},
			want: []BreakingChange{
				{Type: "PARAMETER_TYPE_CHANGED", Severity: SeverityHigh},
			},
		},
		{
			name: "new required parameter",
			base: &MCPState{
				Tools: map[string]ToolDef{
					"tool1": {
						Name: "tool1",
						InputSchema: map[string]any{
							"properties": map[string]any{
								"param1": map[string]any{"type": "string"},
							},
							"required": []any{},
						},
					},
				},
			},
			current: &MCPState{
				Tools: map[string]ToolDef{
					"tool1": {
						Name: "tool1",
						InputSchema: map[string]any{
							"properties": map[string]any{
								"param1": map[string]any{"type": "string"},
								"param2": map[string]any{"type": "string"},
							},
							"required": []any{"param2"},
						},
					},
				},
			},
			want: []BreakingChange{
				{Type: "NEW_REQUIRED_PARAMETER", Severity: SeverityHigh},
			},
		},
		{
			name: "resource uri removed",
			base: &MCPState{
				Resources: map[string]ResourceDef{
					"file:///res": {URI: "file:///res"},
				},
			},
			current: &MCPState{
				Resources: map[string]ResourceDef{},
			},
			want: []BreakingChange{
				{Type: "RESOURCE_URI_REMOVED", Severity: SeverityHigh},
			},
		},
		{
			name: "resource mime type changed",
			base: &MCPState{
				Resources: map[string]ResourceDef{
					"file:///res": {URI: "file:///res", MimeType: "application/json"},
				},
			},
			current: &MCPState{
				Resources: map[string]ResourceDef{
					"file:///res": {URI: "file:///res", MimeType: "text/plain"},
				},
			},
			want: []BreakingChange{
				{Type: "RESOURCE_MIME_TYPE_CHANGED", Severity: SeverityHigh},
			},
		},
		{
			name: "prompt removed",
			base: &MCPState{
				Prompts: map[string]PromptDef{
					"prompt1": {Name: "prompt1"},
				},
			},
			current: &MCPState{
				Prompts: map[string]PromptDef{},
			},
			want: []BreakingChange{
				{Type: "PROMPT_REMOVED", Severity: SeverityHigh},
			},
		},
		{
			name: "prompt required argument added",
			base: &MCPState{
				Prompts: map[string]PromptDef{
					"prompt1": {
						Name: "prompt1",
						Arguments: []PromptArgument{
							{Name: "arg1", Required: false},
						},
					},
				},
			},
			current: &MCPState{
				Prompts: map[string]PromptDef{
					"prompt1": {
						Name: "prompt1",
						Arguments: []PromptArgument{
							{Name: "arg1", Required: false},
							{Name: "arg2", Required: true},
						},
					},
				},
			},
			want: []BreakingChange{
				{Type: "PROMPT_REQUIRED_ARGUMENT_ADDED", Severity: SeverityHigh},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DiffMCPState(tt.base, tt.current)
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d changes, got %d", len(tt.want), len(got))
			}
			for i, w := range tt.want {
				if got[i].Type != w.Type {
					t.Errorf("expected type %q, got %q", w.Type, got[i].Type)
				}
				if got[i].Severity != w.Severity {
					t.Errorf("expected severity %q, got %q", w.Severity, got[i].Severity)
				}
			}
		})
	}
}

// TestHelperProcess isn't a real test; it's a helper process for TestCaptureMCPState.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	// We are mimicking an MCP server responding to stdio JSON-RPC requests
	d := json.NewDecoder(os.Stdin)
	e := json.NewEncoder(os.Stdout)

	for {
		var req JSONRPCRequest
		if err := d.Decode(&req); err != nil {
			break
		}

		if req.Method == "initialize" {
			initResp := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: json.RawMessage(`{
					"protocolVersion": "2024-11-05",
					"serverInfo": {"name": "test-mcp", "version": "1.0"},
					"capabilities": {"tools": {}}
				}`),
			}
			e.Encode(initResp)
		} else if req.Method == "notifications/initialized" {
			// Do nothing
		} else if req.Method == "tools/list" {
			toolsResp := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: json.RawMessage(`{
					"tools": [
						{
							"name": "test_tool",
							"inputSchema": {
								"type": "object",
								"properties": {
									"arg1": {"type": "string"}
								}
							}
						}
					]
				}`),
			}
			e.Encode(toolsResp)
		} else if req.Method == "resources/list" {
			resResp := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: json.RawMessage(`{
					"resources": [
						{
							"uri": "file:///test",
							"name": "test_res",
							"mimeType": "application/json"
						}
					]
				}`),
			}
			e.Encode(resResp)
		} else if req.Method == "prompts/list" {
			promptResp := JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      req.ID,
				Result: json.RawMessage(`{
					"prompts": [
						{
							"name": "test_prompt",
							"arguments": [
								{
									"name": "arg1",
									"required": true
								}
							]
						}
					]
				}`),
			}
			e.Encode(promptResp)
			break // exit after prompts/list to finish the command gracefully
		}
	}
}

func TestCaptureMCPState(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Use t.TempDir() for the mock executable to avoid permissions/cleanup issues
	tmpDir := t.TempDir()
	mockPath := tmpDir + "/mock_mcp"

	cmd := exec.Command("go", "build", "-o", mockPath, "../../testdata/mock_mcp.go")
	if err := cmd.Run(); err != nil {
		// Output the go command's stderr for debugging
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("Failed to build mock_mcp: %v\nstderr: %s", err, string(ee.Stderr))
		}
		t.Fatalf("Failed to build mock_mcp: %v", err)
	}

	state, err := CaptureMCPState(ctx, mockPath, []string{})
	if err != nil {
		t.Fatalf("CaptureMCPState failed: %v", err)
	}

	if len(state.Tools) != 1 || state.Tools["test_tool"].Name != "test_tool" {
		t.Errorf("unexpected tools: %v", state.Tools)
	}
	if len(state.Resources) != 1 || state.Resources["file:///test"].Name != "test_res" {
		t.Errorf("unexpected resources: %v", state.Resources)
	}
	if len(state.Prompts) != 1 || state.Prompts["test_prompt"].Name != "test_prompt" {
		t.Errorf("unexpected prompts: %v", state.Prompts)
	}
}
