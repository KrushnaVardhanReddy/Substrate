package main

import (
	"encoding/json"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
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
