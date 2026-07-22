package ai

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestRunArchitect(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		apiKey         string
		baseURL        string
		model          string
		mockResponse   string
		expectedError  string
		expectedOutput string
		expectedFile   string
	}{
		{
			name:          "Empty input",
			input:         "\n",
			apiKey:        "test-key",
			expectedError: "input cannot be empty",
		},
		{
			name:          "Missing API key without fallback URL",
			input:         "test\n",
			apiKey:        "",
			baseURL:       "http://some-url",
			expectedError: "SUBSTRATE_AI_API_KEY environment variable is required",
		},
		{
			name:           "Missing API key allowed for Ollama",
			input:          "test\n",
			apiKey:         "",
			baseURL:        "MOCK_SERVER", // We use mock server for this to avoid actual network call
			mockResponse:   "data: {\"choices\": [{\"delta\": {\"content\": \"```yaml\\nopenapi: 3.0.0\\n```\"}}]}\n\ndata: [DONE]\n\n",
			expectedError:  "",
			expectedFile:   "openapi: 3.0.0\n",
			expectedOutput: "openapi: 3.0.0",
		},
		{
			name:           "Bedrock mock server",
			input:          "users api\n",
			apiKey:         "test-key",
			baseURL:        "MOCK_SERVER", // Will be replaced in test setup
			mockResponse:   "```yaml\nopenapi: 3.0.0\n```",
			expectedError:  "",
			expectedFile:   "openapi: 3.0.0\n",
			expectedOutput: "openapi: 3.0.0",
		},
		{
			name:           "Fallback triggered when baseURL is unset",
			input:          "blog api\n",
			apiKey:         "",
			baseURL:        "",
			expectedError:  "",
			expectedFile:   FallbackOpenAPI,
			expectedOutput: "Using deterministic fallback response",
		},
		{
			name:           "Success with mock server",
			input:          "users api\n",
			apiKey:         "test-key",
			baseURL:        "MOCK_SERVER", // Will be replaced in test setup
			mockResponse:   "data: {\"choices\": [{\"delta\": {\"content\": \"```yaml\\nopenapi: 3.0.0\\n```\"}}]}\n\ndata: [DONE]\n\n",
			expectedError:  "",
			expectedFile:   "openapi: 3.0.0\n",
			expectedOutput: "openapi: 3.0.0",
		},
		{
			name:          "Closed input unexpectedly",
			input:         "",
			apiKey:        "test-key",
			expectedError: "input closed unexpectedly",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SUBSTRATE_AI_API_KEY", tt.apiKey)
			t.Setenv("SUBSTRATE_AI_MODEL", tt.model)

			if tt.name == "Missing API key allowed for Ollama" {
				t.Setenv("SUBSTRATE_AI_PROVIDER", "ollama")
			} else if tt.name == "Bedrock mock server" {
				t.Setenv("SUBSTRATE_AI_PROVIDER", "bedrock")
			} else {
				t.Setenv("SUBSTRATE_AI_PROVIDER", "openai")
			}

			var mockServer *httptest.Server
			if tt.baseURL == "MOCK_SERVER" {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "text/event-stream")
					fmt.Fprint(w, tt.mockResponse)
				}))
				defer mockServer.Close()
				t.Setenv("SUBSTRATE_AI_BASE_URL", mockServer.URL+"/v1")
			} else {
				t.Setenv("SUBSTRATE_AI_BASE_URL", tt.baseURL)
			}

			// Change directory to TempDir to avoid side effects
			tempDir := t.TempDir()
			origDir, _ := os.Getwd()
			os.Chdir(tempDir)
			defer os.Chdir(origDir)

			in := strings.NewReader(tt.input)
			out := &bytes.Buffer{}

			err := RunArchitect(in, out)

			if tt.expectedError != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.expectedError)
				}
				if !strings.Contains(err.Error(), tt.expectedError) {
					t.Errorf("expected error containing %q, got %q", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				outputStr := out.String()
				if tt.expectedOutput != "" && !strings.Contains(outputStr, tt.expectedOutput) {
					t.Errorf("expected output to contain %q, got %q", tt.expectedOutput, outputStr)
				}

				if tt.expectedFile != "" {
					content, readErr := os.ReadFile("openapi.yaml")
					if readErr != nil {
						t.Fatalf("failed to read generated file: %v", readErr)
					}
					if string(content) != tt.expectedFile {
						t.Errorf("expected file content %q, got %q", tt.expectedFile, string(content))
					}
				}
			}
		})
	}
}

func TestExtractYAML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "No markdown",
			input:    "openapi: 3.0.0",
			expected: "openapi: 3.0.0\n",
		},
		{
			name:     "Markdown with yaml",
			input:    "```yaml\nopenapi: 3.0.0\n```",
			expected: "openapi: 3.0.0\n",
		},
		{
			name:     "Markdown without yaml",
			input:    "```\nopenapi: 3.0.0\n```",
			expected: "openapi: 3.0.0\n",
		},
		{
			name:     "Whitespace around",
			input:    "   ```yaml\nopenapi: 3.0.0\n```   ",
			expected: "openapi: 3.0.0\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractYAML(tt.input)
			if result != tt.expected {
				t.Errorf("extractYAML(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}
