package init

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(dir string)
		opts           InitOptions
		expectedFiles  []string
		expectedOutput []string
		expectedYAML   map[string]string // FilePath -> sub-strings that must be present
	}{
		{
			name: "No existing files, no flags",
			opts: InitOptions{SchemaType: "openapi", Branch: "main"},
			expectedFiles: []string{
				"substrate.yaml",
				".github/workflows/substrate.yml",
			},
			expectedOutput: []string{
				"⚠️  No OpenAPI spec found. Set spec_path in substrate.yaml to the correct path.",
				"✅ Substrate initialized!",
				"  substrate.yaml",
				"  .github/workflows/substrate.yml",
			},
			expectedYAML: map[string]string{
				"substrate.yaml":                  "spec_path: openapi.yaml",
				".github/workflows/substrate.yml": "base_schema: openapi.yaml",
			},
		},
		{
			name: "No existing files, --spec api/openapi.yaml",
			opts: InitOptions{Spec: "api/openapi.yaml", SchemaType: "openapi", Branch: "main"},
			expectedFiles: []string{
				"substrate.yaml",
				".github/workflows/substrate.yml",
			},
			expectedOutput: []string{
				"✅ Substrate initialized!",
			},
			expectedYAML: map[string]string{
				"substrate.yaml":                  "spec_path: api/openapi.yaml",
				".github/workflows/substrate.yml": "base_schema: api/openapi.yaml",
			},
		},
		{
			name: "Auto-detect spec",
			setup: func(dir string) {
				os.WriteFile(filepath.Join(dir, "openapi.yaml"), []byte(""), 0644)
			},
			opts: InitOptions{SchemaType: "openapi", Branch: "main"},
			expectedFiles: []string{
				"substrate.yaml",
				".github/workflows/substrate.yml",
			},
			expectedOutput: []string{
				"✅ Substrate initialized!",
			},
			expectedYAML: map[string]string{
				"substrate.yaml":                  "spec_path: openapi.yaml",
				".github/workflows/substrate.yml": "base_schema: openapi.yaml",
			},
		},
		{
			name: "Files already exist, no --force",
			setup: func(dir string) {
				os.WriteFile(filepath.Join(dir, "substrate.yaml"), []byte("existing"), 0644)
				os.MkdirAll(filepath.Join(dir, ".github", "workflows"), 0755)
				os.WriteFile(filepath.Join(dir, ".github", "workflows", "substrate.yml"), []byte("existing"), 0644)
			},
			opts: InitOptions{SchemaType: "openapi", Branch: "main"},
			expectedFiles: []string{
				"substrate.yaml",
				".github/workflows/substrate.yml",
			},
			expectedOutput: []string{
				"⚠️  substrate.yaml already exists. Use --force to overwrite.",
				"⚠️  .github/workflows/substrate.yml already exists. Use --force to overwrite.",
			},
			expectedYAML: map[string]string{
				"substrate.yaml":                  "existing",
				".github/workflows/substrate.yml": "existing",
			},
		},
		{
			name: "Files already exist, --force",
			setup: func(dir string) {
				os.WriteFile(filepath.Join(dir, "substrate.yaml"), []byte("existing"), 0644)
				os.MkdirAll(filepath.Join(dir, ".github", "workflows"), 0755)
				os.WriteFile(filepath.Join(dir, ".github", "workflows", "substrate.yml"), []byte("existing"), 0644)
			},
			opts: InitOptions{Force: true, SchemaType: "openapi", Branch: "main"},
			expectedFiles: []string{
				"substrate.yaml",
				".github/workflows/substrate.yml",
			},
			expectedOutput: []string{
				"✅ Substrate initialized!",
			},
			expectedYAML: map[string]string{
				"substrate.yaml":                  "spec_path: openapi.yaml",
				".github/workflows/substrate.yml": "base_schema: openapi.yaml",
			},
		},
		{
			name: "--no-workflow flag",
			opts: InitOptions{NoWorkflow: true, SchemaType: "openapi", Branch: "main"},
			expectedFiles: []string{
				"substrate.yaml",
			},
			expectedOutput: []string{
				"✅ Substrate initialized!",
				"  substrate.yaml",
			},
			expectedYAML: map[string]string{
				"substrate.yaml": "spec_path: openapi.yaml",
			},
		},
		{
			name: "--no-config flag",
			opts: InitOptions{NoConfig: true, SchemaType: "openapi", Branch: "main"},
			expectedFiles: []string{
				".github/workflows/substrate.yml",
			},
			expectedOutput: []string{
				"✅ Substrate initialized!",
				"  .github/workflows/substrate.yml",
			},
			expectedYAML: map[string]string{
				".github/workflows/substrate.yml": "base_schema: openapi.yaml",
			},
		},
		{
			name: "--service my-api",
			opts: InitOptions{Service: "my-api", SchemaType: "openapi", Branch: "main"},
			expectedFiles: []string{
				"substrate.yaml",
				".github/workflows/substrate.yml",
			},
			expectedOutput: []string{
				"✅ Substrate initialized!",
			},
			expectedYAML: map[string]string{
				"substrate.yaml": "service: my-api",
			},
		},
		{
			name: "--branch develop",
			opts: InitOptions{Branch: "develop", SchemaType: "openapi"},
			expectedFiles: []string{
				"substrate.yaml",
				".github/workflows/substrate.yml",
			},
			expectedOutput: []string{
				"✅ Substrate initialized!",
			},
			expectedYAML: map[string]string{
				".github/workflows/substrate.yml": "branches: [develop]",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			originalWD, err := os.Getwd()
			if err != nil {
				t.Fatalf("failed to get original wd: %v", err)
			}
			err = os.Chdir(dir)
			if err != nil {
				t.Fatalf("failed to chdir to %s: %v", dir, err)
			}
			defer os.Chdir(originalWD)

			if tc.setup != nil {
				tc.setup(dir)
			}

			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err = Init(tc.opts)
			if err != nil {
				t.Fatalf("Init() unexpected error: %v", err)
			}

			w.Close()
			var buf bytes.Buffer
			io.Copy(&buf, r)
			os.Stdout = oldStdout

			output := buf.String()
			for _, expectedMsg := range tc.expectedOutput {
				if !strings.Contains(output, expectedMsg) {
					t.Errorf("expected output to contain %q, got: %s", expectedMsg, output)
				}
			}

			if tc.name == "--no-workflow flag" && strings.Contains(output, ".github/workflows/substrate.yml") {
				t.Errorf("output should not contain .github/workflows/substrate.yml")
			}
			if tc.name == "--no-config flag" && strings.Contains(output, "  substrate.yaml") {
				t.Errorf("output should not contain substrate.yaml")
			}

			// Auto-detection message checks
			if tc.name == "No existing files, no flags" && !strings.Contains(output, "⚠️  No OpenAPI spec found.") {
				t.Errorf("expected output to contain missing spec warning, got: %s", output)
			}
			if tc.name == "Auto-detect spec" && strings.Contains(output, "⚠️  No OpenAPI spec found.") {
				t.Errorf("unexpected missing spec warning in output: %s", output)
			}
			if tc.name == "No existing files, --spec api/openapi.yaml" && strings.Contains(output, "⚠️  No OpenAPI spec found.") {
				t.Errorf("unexpected missing spec warning in output: %s", output)
			}

			for _, file := range tc.expectedFiles {
				if !fileExists(filepath.Join(dir, file)) {
					t.Errorf("expected file %q to be created, but it was not", file)
				}
			}

			if tc.opts.NoWorkflow && fileExists(filepath.Join(dir, ".github/workflows/substrate.yml")) {
				t.Errorf("expected .github/workflows/substrate.yml to not be created")
			}
			if tc.opts.NoConfig && fileExists(filepath.Join(dir, "substrate.yaml")) {
				t.Errorf("expected substrate.yaml to not be created")
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
