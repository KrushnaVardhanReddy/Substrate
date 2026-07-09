package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecutionModes(t *testing.T) {
	binaryPath, _ := filepath.Abs("substrate-test")
	buildCmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary for tests: %v", err)
	}
	defer os.Remove(binaryPath)

	tests := []struct {
		name           string
		args           []string
		expectedCode   int
		expectedOutput []string
	}{
		{
			name:           "Strict mode with breaking change -> Expect exit code 2",
			args:           []string{"diff", "testdata/base.yaml", "testdata/rev_breaking.yaml", "--mode=strict"},
			expectedCode:   2,
			expectedOutput: []string{},
		},
		{
			name:           "Legacy mode with breaking change -> Expect exit code 0 and warning message",
			args:           []string{"diff", "testdata/base.yaml", "testdata/rev_breaking.yaml", "--mode=legacy", "--format=text"},
			expectedCode:   0,
			expectedOutput: []string{"⚠️ LEGACY MODE: Breaking changes detected, but merge is not blocked."},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()

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

			outputStr := string(output)
			for _, expected := range tt.expectedOutput {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput:\n%s", expected, outputStr)
				}
			}
		})
	}
}
