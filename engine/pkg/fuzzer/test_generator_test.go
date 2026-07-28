package fuzzer

import (
	"os"
	"strings"
	"testing"
)

func TestGenerateTests(t *testing.T) {
	// Create a temporary OpenAPI spec for testing
	specContent := `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      requestBody:
        required: true
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
        '201':
          description: Created
`
	tmpfile, err := os.CreateTemp("", "spec-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(specContent)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name     string
		specPath string
		baseURL  string
		wantErr  bool
	}{
		{
			name:     "valid spec",
			specPath: tmpfile.Name(),
			baseURL:  "http://api.example.com",
			wantErr:  false,
		},
		{
			name:     "invalid spec path",
			specPath: "non_existent.yaml",
			baseURL:  "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := GenerateTests(tt.specPath, tt.baseURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateTests() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// verify output contains expected test structure
				if !strings.Contains(output, "package tests") {
					t.Errorf("output does not contain package tests")
				}
				if !strings.Contains(output, "func TestGeneratedFuzzing") {
					t.Errorf("output does not contain TestGeneratedFuzzing")
				}
				// verify it generated tests for constraints
				if !strings.Contains(output, "missing required name") {
					t.Errorf("output does not contain required test case")
				}
				if !strings.Contains(output, "exceeds maxLength") {
					t.Errorf("output does not contain maxLength test case")
				}
				if !strings.Contains(output, "below minimum") {
					t.Errorf("output does not contain minimum test case")
				}

				// Verify it generated security payloads
				if !strings.Contains(output, "SQL injection") {
					t.Errorf("output does not contain SQL injection test case")
				}
				if !strings.Contains(output, "Path traversal") {
					t.Errorf("output does not contain Path traversal test case")
				}
				if !strings.Contains(output, "Null byte injection") {
					t.Errorf("output does not contain Null byte injection test case")
				}
				if !strings.Contains(output, "Extremely long string") {
					t.Errorf("output does not contain Extremely long string test case")
				}
			}
		})
	}
}
