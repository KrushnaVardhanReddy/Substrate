package fuzzer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateTests(t *testing.T) {
	// Create a temporary OpenAPI spec file
	schemaContent := `openapi: "3.0.0"
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
                  maxLength: 10
                age:
                  type: integer
                  minimum: 18
      responses:
        "201":
          description: Created
`
	tmpDir := t.TempDir()
	specPath := filepath.Join(tmpDir, "openapi.yaml")
	err := os.WriteFile(specPath, []byte(schemaContent), 0644)
	if err != nil {
		t.Fatalf("failed to write spec file: %v", err)
	}

	tests := []struct {
		name          string
		specPath      string
		baseURL       string
		expectStrings []string
	}{
		{
			name:     "Valid OpenAPI spec with constraints",
			specPath: specPath,
			baseURL:  "http://example.com/api",
			expectStrings: []string{
				`e := httpexpect.Default(t, "http://example.com/api")`,
				`t.Run("POST /users missing required name"`,
				`t.Run("POST /users name exceeds maxLength"`,
				`t.Run("POST /users age below minimum"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := GenerateTests(tt.specPath, tt.baseURL)
			if err != nil {
				t.Fatalf("GenerateTests returned error: %v", err)
			}
			for _, exp := range tt.expectStrings {
				if !strings.Contains(output, exp) {
					t.Errorf("Expected output to contain:\n%s\nGot:\n%s", exp, output)
				}
			}
		})
	}
}
