package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMCPServerE2E(t *testing.T) {
	// Build the CLI binary
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "substrate-mcp")

	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := buildCmd.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build binary: %v\nOutput: %s", err, out)
	}

	// Prepare JSON-RPC input
	requests := []string{
		`{"jsonrpc": "2.0", "id": 1, "method": "initialize"}`,
		`{"jsonrpc": "2.0", "id": 2, "method": "tools/list"}`,
		`{"jsonrpc": "2.0", "id": 3, "method": "tools/call", "params": {"name": "get_substrate_docs", "arguments": {}}}`,
	}
	inputStr := strings.Join(requests, "\n") + "\n"

	// Run the built binary
	cmd := exec.Command(binPath)
	cmd.Stdin = bytes.NewBufferString(inputStr)
	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf

	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run binary: %v", err)
	}

	// Read output lines
	outputLines := strings.Split(strings.TrimSpace(outBuf.String()), "\n")
	if len(outputLines) != 3 {
		t.Fatalf("Expected 3 responses, got %d. Output:\n%s", len(outputLines), outBuf.String())
	}

	// Verify first response (initialize)
	var resp1 map[string]any
	if err := json.Unmarshal([]byte(outputLines[0]), &resp1); err != nil {
		t.Fatalf("Failed to parse first response: %v", err)
	}
	if resp1["id"].(float64) != 1 {
		t.Errorf("Expected id 1 for first response, got %v", resp1["id"])
	}

	// Verify second response (tools/list)
	var resp2 map[string]any
	if err := json.Unmarshal([]byte(outputLines[1]), &resp2); err != nil {
		t.Fatalf("Failed to parse second response: %v", err)
	}
	if resp2["id"].(float64) != 2 {
		t.Errorf("Expected id 2 for second response, got %v", resp2["id"])
	}
	result2, ok := resp2["result"].(map[string]any)
	if !ok {
		t.Fatalf("Expected result object in second response")
	}
	tools, ok := result2["tools"].([]any)
	if !ok || len(tools) != 7 {
		t.Errorf("Expected 7 tools, got %d", len(tools))
	}

	// Verify third response (tools/call)
	var resp3 map[string]any
	if err := json.Unmarshal([]byte(outputLines[2]), &resp3); err != nil {
		t.Fatalf("Failed to parse third response: %v", err)
	}
	if resp3["id"].(float64) != 3 {
		t.Errorf("Expected id 3 for third response, got %v", resp3["id"])
	}
	result3, ok := resp3["result"].(map[string]any)
	if !ok {
		t.Fatalf("Expected result object in third response")
	}
	content, ok := result3["content"].([]any)
	if !ok || len(content) == 0 {
		t.Fatalf("Expected content array in third response")
	}
	contentObj := content[0].(map[string]any)
	text, ok := contentObj["text"].(string)
	if !ok {
		t.Fatalf("Expected text string in content")
	}
	if !strings.Contains(text, "Substrate Configuration Guide") {
		t.Errorf("Expected 'Substrate Configuration Guide' in response, got %s", text)
	}
}
