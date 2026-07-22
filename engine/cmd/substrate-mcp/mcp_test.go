package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// ImpactResponse represents the expected JSON response from the API
type ImpactResponse struct {
	Provider      string   `json:"provider"`
	RiskScore     int      `json:"risk_score"`
	ImpactedRepos []string `json:"impacted_repos"`
}

func TestMCPGetBlastRadius(t *testing.T) {
	// Build the binary dynamically for the test
	cmdBuild := exec.Command("go", "build", "-o", "test-substrate-mcp")
	if err := cmdBuild.Run(); err != nil {
		t.Fatalf("failed to build test binary: %v", err)
	}
	defer os.Remove("test-substrate-mcp") // cleanup

	// Mock Substrate Registry API
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/impact/myorg/backend-api" {
			resp := ImpactResponse{
				Provider:      "myorg/backend-api",
				RiskScore:     2,
				ImpactedRepos: []string{"myorg/consumer-a", "myorg/consumer-b"},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer apiServer.Close()

	os.Setenv("REGISTRY_API_URL", apiServer.URL)
	defer os.Unsetenv("REGISTRY_API_URL")

	// Start test-substrate-mcp
	cmd := exec.Command("./test-substrate-mcp")

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	defer cmd.Process.Kill()

	// Send an MCP JSON-RPC call to `get_blast_radius`
	req := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"get_blast_radius","arguments":{"repo":"myorg/backend-api"}}}` + "\n"
	if _, err := io.WriteString(stdin, req); err != nil {
		t.Fatal(err)
	}

	// Read response
	buf := make([]byte, 2048)
	n, err := stdout.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}

	respStr := string(buf[:n])

	if !strings.Contains(respStr, "myorg/consumer-a") || !strings.Contains(respStr, "risk_score") {
		t.Errorf("unexpected mcp response: %s", respStr)
	}
}
