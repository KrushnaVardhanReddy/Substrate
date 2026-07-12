package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestE2EGenerateTests(t *testing.T) {
	// Build the CLI binary
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "substrate")

	buildCmd := exec.Command("go", "build", "-o", binPath, ".")
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build CLI: %v", err)
	}

	// Create test OpenAPI spec
	specContent := `openapi: "3.0.0"
info:
  title: Test API
  version: "1.0.0"
paths:
  /users:
    post:
      summary: Create user
      requestBody:
        content:
          application/json:
            schema:
              type: object
              required:
                - name
              properties:
                name:
                  type: string
                  maxLength: 50
                age:
                  type: integer
                  minimum: 18
      responses:
        "201":
          description: Created
`
	specPath := filepath.Join(tmpDir, "openapi.yaml")
	if err := os.WriteFile(specPath, []byte(specContent), 0644); err != nil {
		t.Fatalf("Failed to write test spec: %v", err)
	}

	// Run the generate-tests command
	cmd := exec.Command(binPath, "generate-tests", specPath, "--url", "http://localhost:8080")
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()

	if err != nil {
		t.Fatalf("generate-tests failed: %v\nstderr: %s", err, stderr.String())
	}

	output := out.String()
	expectedStrings := []string{
		`package tests`,
		`github.com/gavv/httpexpect/v2`,
		`e := httpexpect.Default(t, "http://localhost:8080")`,
		`t.Run("POST /users missing required name"`,
		`t.Run("POST /users name exceeds maxLength"`,
		`t.Run("POST /users age below minimum"`,
	}

	for _, exp := range expectedStrings {
		if !strings.Contains(output, exp) {
			t.Errorf("Expected output to contain: %s\nGot output:\n%s", exp, output)
		}
	}
}
