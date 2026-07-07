package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitE2E(t *testing.T) {
	// Build the binary
	binaryPath := filepath.Join(t.TempDir(), "substrate")
	cmd := exec.Command("go", "build", "-o", binaryPath, ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	tests := []struct {
		name           string
		args           []string
		setup          func(dir string)
		expectedCode   int
		expectedFiles  []string
		expectedOutput []string
		expectedYAML   map[string]string // FilePath -> expected string content
	}{
		{
			name:         "substrate init in empty dir → creates both files, exit 0",
			args:         []string{"init"},
			expectedCode: 0,
			expectedFiles: []string{
				"substrate.yaml",
				".github/workflows/substrate.yml",
			},
			expectedOutput: []string{
				"✅ Substrate initialized!",
				"⚠️  No OpenAPI spec found.",
			},
		},
		{
			name:         "substrate init --no-workflow → creates only substrate.yaml, exit 0",
			args:         []string{"init", "--no-workflow"},
			expectedCode: 0,
			expectedFiles: []string{
				"substrate.yaml",
			},
			expectedOutput: []string{
				"✅ Substrate initialized!",
			},
		},
		{
			name:         "substrate init --service my-api → service: my-api in substrate.yaml",
			args:         []string{"init", "--service", "my-api"},
			expectedCode: 0,
			expectedFiles: []string{
				"substrate.yaml",
			},
			expectedOutput: []string{
				"✅ Substrate initialized!",
			},
			expectedYAML: map[string]string{
				"substrate.yaml": "service: my-api",
			},
		},
		{
			name: "substrate init twice without --force → exit 0, prints warning, files unchanged",
			args: []string{"init"},
			setup: func(dir string) {
				os.WriteFile(filepath.Join(dir, "substrate.yaml"), []byte("existing config"), 0644)
			},
			expectedCode: 0,
			expectedOutput: []string{
				"⚠️  substrate.yaml already exists. Use --force to overwrite.",
			},
			expectedYAML: map[string]string{
				"substrate.yaml": "existing config",
			},
		},
		{
			name: "substrate init --force → overwrites existing files",
			args: []string{"init", "--force"},
			setup: func(dir string) {
				os.WriteFile(filepath.Join(dir, "substrate.yaml"), []byte("existing config"), 0644)
			},
			expectedCode: 0,
			expectedOutput: []string{
				"✅ Substrate initialized!",
			},
			expectedYAML: map[string]string{
				"substrate.yaml": "service:",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()

			if tc.setup != nil {
				tc.setup(dir)
			}

			cmd := exec.Command(binaryPath, tc.args...)
			cmd.Dir = dir
			var outBuf, errBuf bytes.Buffer
			cmd.Stdout = &outBuf
			cmd.Stderr = &errBuf

			err := cmd.Run()

			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					t.Fatalf("failed to run command: %v", err)
				}
			}

			if exitCode != tc.expectedCode {
				t.Errorf("expected exit code %d, got %d (stdout: %s, stderr: %s)", tc.expectedCode, exitCode, outBuf.String(), errBuf.String())
			}

			output := outBuf.String() + errBuf.String()
			for _, expectedMsg := range tc.expectedOutput {
				if !strings.Contains(output, expectedMsg) {
					t.Errorf("expected output to contain %q, got: %s", expectedMsg, output)
				}
			}

			if tc.name == "substrate init --no-workflow → creates only substrate.yaml, exit 0" {
				if _, err := os.Stat(filepath.Join(dir, ".github/workflows/substrate.yml")); err == nil {
					t.Errorf("expected .github/workflows/substrate.yml to not exist")
				}
			}

			for _, file := range tc.expectedFiles {
				if _, err := os.Stat(filepath.Join(dir, file)); err != nil {
					t.Errorf("expected file %q to exist, but it does not", file)
				}
			}

			for file, expectedContent := range tc.expectedYAML {
				content, err := os.ReadFile(filepath.Join(dir, file))
				if err != nil {
					t.Fatalf("failed to read expected file %q: %v", file, err)
				}
				if !strings.Contains(string(content), expectedContent) {
					t.Errorf("expected file %q to contain %q, got:\n%s", file, expectedContent, string(content))
				}
			}
		})
	}
}
