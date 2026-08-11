// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
)

var binaryPath string

func TestMain(m *testing.M) {
	// Build the binary
	binaryPath, _ = filepath.Abs("substrate-test")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if err := buildCmd.Run(); err != nil {
		os.Stderr.WriteString("Failed to build binary for tests: " + err.Error() + "\n")
		os.Exit(1)
	}

	// Run tests
	exitCode := m.Run()

	// Clean up
	os.Remove(binaryPath)

	os.Exit(exitCode)
}

func TestE2ECli(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectedCode   int
		expectedOutput []string
	}{
		{
			name:           "Safe change -> Expect exit code 0",
			args:           []string{"diff", "testdata/base.yaml", "testdata/base.yaml", "--format=json"},
			expectedCode:   0,
			expectedOutput: []string{`"overall_severity": "NO_CHANGES"`},
		},
		{
			name:           "Warning change -> Expect exit code 1",
			args:           []string{"diff", "testdata/base.yaml", "testdata/rev_warning.yaml", "--format=text"},
			expectedCode:   1,
			expectedOutput: []string{"WARNINGS"},
		},
		{
			name:           "Breaking change -> Expect exit code 2",
			args:           []string{"diff", "testdata/base.yaml", "testdata/rev_breaking.yaml", "--format=changelog"},
			expectedCode:   2,
			expectedOutput: []string{"Breaking Changes", "ENDPOINT_REMOVED"},
		},
		{
			name:           "Breaking change WITH valid override -> Expect exit code 0 or 1",
			args:           []string{"diff", "testdata/base.yaml", "testdata/rev_breaking.yaml", "--config=testdata/substrate_valid.yaml"},
			expectedCode:   0, // Based on specs, moving a single breaking change with override either leaves warnings (exit 1) or safe (exit 0). Here we will make it 0.
			expectedOutput: []string{`"overall_severity": "SAFE"`},
		},
		{
			name:           "Invalid OpenAPI spec -> Expect exit code 3",
			args:           []string{"diff", "testdata/invalid_spec.yaml", "testdata/base.yaml"},
			expectedCode:   3,
			expectedOutput: []string{"Error:"},
		},
		{
			name:           "Missing substrate.yaml required fields -> Expect exit code 3",
			args:           []string{"diff", "testdata/base.yaml", "testdata/rev_breaking.yaml", "--config=testdata/substrate_invalid.yaml"},
			expectedCode:   3,
			expectedOutput: []string{"substrate.yaml: 'service' is required"},
		},
		{
			name:           "AIML Breaking change -> Expect exit code 2",
			args:           []string{"diff", "testdata/aiml/base.yaml", "testdata/aiml/head_breaking.yaml", "--schema-type=ai-model", "--format=json"},
			expectedCode:   2,
			expectedOutput: []string{`"overall_severity": "BREAKING"`, `"AIML_INPUT_REMOVED"`},
		},
		{
			name:           "Custom rule failure -> Expect exit code 2",
			args:           []string{"diff", "testdata/base.yaml", "testdata/rev_custom_rule_fail.yaml", "--config=testdata/substrate_custom_rules.yaml", "--format=json"},
			expectedCode:   2,
			expectedOutput: []string{`"rule_id": "REQUIRE_API_VERSION_2"`, `"overall_severity": "BREAKING"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()

			// Check exit code
			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					t.Fatalf("Failed to run command: %v\nOutput: %s", err, string(output))
				}
			}

			if exitCode != tt.expectedCode {
				t.Errorf("Expected exit code %d, got %d.\nCommand: %s\nOutput:\n%s", tt.expectedCode, exitCode, strings.Join(cmd.Args, " "), string(output))
			}

			// Check expected output
			outputStr := string(output)
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", expected, outputStr)
				}
			}
		})
	}
}

func TestCrossRepoCheckE2E(t *testing.T) {
	cmd := exec.Command(binaryPath, "diff", "testdata/cross-repo/consumer_snapshot.yaml", "testdata/cross-repo/provider_head.yaml", "--schema-type", "openapi", "--format", "json")
	output, err := cmd.CombinedOutput()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			t.Fatalf("Failed to run command: %v\nOutput: %s", err, string(output))
		}
	}

	if exitCode != 2 {
		t.Errorf("Expected exit code 2, got %d.\nCommand: %s\nOutput:\n%s", exitCode, strings.Join(cmd.Args, " "), string(output))
	}

	var rep report.DiffReport
	if err := json.Unmarshal(output, &rep); err != nil {
		t.Fatalf("Failed to parse output as JSON: %v\nOutput: %s", err, string(output))
	}

	if rep.Summary.BreakingCount != 1 {
		t.Errorf("Expected Summary.BreakingCount to be 1, got %d", rep.Summary.BreakingCount)
	}

	if len(rep.BreakingChanges) == 0 {
		t.Fatalf("Expected at least 1 BreakingChange, got 0")
	}

	if rep.BreakingChanges[0].RuleID != "ENDPOINT_REMOVED" {
		t.Errorf("Expected BreakingChanges[0].RuleID to be ENDPOINT_REMOVED, got %s", rep.BreakingChanges[0].RuleID)
	}

	if rep.BreakingChanges[0].Path != "GET /users" {
		t.Errorf("Expected BreakingChanges[0].Path to be 'GET /users', got '%s'", rep.BreakingChanges[0].Path)
	}
}

func TestInitDesignE2E(t *testing.T) {
	// Create a mock server for OpenAI streaming
	mockResponse := "data: {\"choices\": [{\"delta\": {\"content\": \"```yaml\\nopenapi: 3.0.0\\ninfo:\\n  title: Mock\\n```\"}}]}\n\ndata: [DONE]\n\n"
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprint(w, mockResponse)
	}))
	defer mockServer.Close()

	// Prepare temporary directory for running the command
	runDir := t.TempDir()

	// Create environment variables
	env := append(os.Environ(),
		"SUBSTRATE_AI_API_KEY=test-key",
		"SUBSTRATE_AI_BASE_URL="+mockServer.URL+"/v1",
	)

	// Run the init command with --design
	cmd := exec.Command(binaryPath, "init", "--design")
	cmd.Dir = runDir
	cmd.Env = env

	// Provide standard input
	cmd.Stdin = strings.NewReader("I want a mock API\n")
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Command failed with error: %v\nStdout: %s\nStderr: %s", err, stdoutBuf.String(), stderrBuf.String())
	}

	stdout := stdoutBuf.String()
	if !strings.Contains(stdout, "openapi: 3.0.0") {
		t.Errorf("Expected stdout to contain 'openapi: 3.0.0', but got: %s", stdout)
	}
	if !strings.Contains(stdout, "Substrate initialized!") {
		t.Errorf("Expected stdout to contain 'Substrate initialized!', but got: %s", stdout)
	}

	// Check if openapi.yaml was generated
	openapiPath := filepath.Join(runDir, "openapi.yaml")
	content, err := os.ReadFile(openapiPath)
	if err != nil {
		t.Fatalf("Failed to read generated openapi.yaml: %v", err)
	}

	expectedYAML := "openapi: 3.0.0\ninfo:\n  title: Mock\n"
	if string(content) != expectedYAML {
		t.Errorf("Expected openapi.yaml content:\n%q\nGot:\n%q", expectedYAML, string(content))
	}

	// Check if substrate.yaml was generated
	configPath := filepath.Join(runDir, "substrate.yaml")
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Errorf("substrate.yaml was not generated")
	}

	// Check if workflow was generated
	workflowPath := filepath.Join(runDir, ".github", "workflows", "substrate.yml")
	if _, err := os.Stat(workflowPath); os.IsNotExist(err) {
		t.Errorf(".github/workflows/substrate.yml was not generated")
	}
}

func TestInitDesignFallbackE2E(t *testing.T) {
	runDir := t.TempDir()

	// Unset SUBSTRATE_AI_BASE_URL to trigger fallback
	var env []string
	for _, e := range os.Environ() {
		if !strings.HasPrefix(e, "SUBSTRATE_AI_BASE_URL=") {
			env = append(env, e)
		}
	}
	env = append(env, "SUBSTRATE_AI_BASE_URL=")

	cmd := exec.Command(binaryPath, "init", "--design")
	cmd.Dir = runDir
	cmd.Env = env

	cmd.Stdin = strings.NewReader("I want a mock API\n")
	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Run()
	if err != nil {
		t.Fatalf("Command failed with error: %v\nStdout: %s\nStderr: %s", err, stdoutBuf.String(), stderrBuf.String())
	}

	stdout := stdoutBuf.String()
	if !strings.Contains(stdout, "Using deterministic fallback response") {
		t.Errorf("Expected stdout to contain 'Using deterministic fallback response', but got: %s", stdout)
	}

	openapiPath := filepath.Join(runDir, "openapi.yaml")
	if _, err := os.Stat(openapiPath); os.IsNotExist(err) {
		t.Errorf("openapi.yaml was not generated")
	}
}
