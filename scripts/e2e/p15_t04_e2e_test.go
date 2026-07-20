package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestLintE2E(t *testing.T) {
	// Need to build the engine CLI first
	engineDir, err := filepath.Abs("../../engine")
	if err != nil {
		t.Fatal(err)
	}

	cliPath := filepath.Join(engineDir, "substrate")
	buildCmd := exec.Command("go", "build", "-o", cliPath, "./cmd/substrate")
	buildCmd.Dir = engineDir
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build substrate CLI: %v", err)
	}
	defer os.Remove(cliPath)

	// Create a dummy schema
	tempDir, err := os.MkdirTemp("", "e2e_lint_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	schemaPath := filepath.Join(tempDir, "schema.yaml")
	err = os.WriteFile(schemaPath, []byte("openapi: 3.0.0\ninfo:\n  title: Mock API\n"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// Run the lint command
	cmd := exec.Command(cliPath, "lint", "--schema", schemaPath)

	// Force deterministic mock response by unsetting AI variables
	cmd.Env = append(os.Environ(),
		"SUBSTRATE_AI_BASE_URL=",
		"OPENAI_API_KEY=",
		"SUBSTRATE_AI_API_KEY=",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("substrate lint failed: %v\nOutput: %s", err, string(output))
	}

	outStr := string(output)
	if !strings.Contains(outStr, "FALLBACK_MODE") {
		t.Errorf("Expected fallback mock response, got: %s", outStr)
	}
	if !strings.Contains(outStr, "\"score\": 90") {
		t.Errorf("Expected score 90 in JSON, got: %s", outStr)
	}
}
