package ai

import (
	"github.com/spf13/viper"
	"strings"
	"testing"
)

// Ensure we test the AI provider switching without making actual HTTP calls
func TestNewAIClient(t *testing.T) {
	tests := []struct {
		name          string
		provider      string
		apiKey        string
		baseURL       string
		expectedError string
	}{
		{
			name:          "Default OpenAI",
			provider:      "",
			apiKey:        "test-key",
			baseURL:       "",
			expectedError: "",
		},
		{
			name:          "Explicit OpenAI",
			provider:      "openai",
			apiKey:        "test-key",
			baseURL:       "",
			expectedError: "",
		},
		{
			name:          "Azure",
			provider:      "azure",
			apiKey:        "test-key",
			baseURL:       "https://mock.azure.com",
			expectedError: "",
		},
		{
			name:          "Bedrock",
			provider:      "bedrock",
			apiKey:        "test-key",
			baseURL:       "https://mock.bedrock.com",
			expectedError: "",
		},
		{
			name:          "Ollama Default",
			provider:      "ollama",
			apiKey:        "",
			baseURL:       "",
			expectedError: "",
		},
		{
			name:          "Ollama Custom",
			provider:      "ollama",
			apiKey:        "",
			baseURL:       "http://custom-ollama:11434/v1",
			expectedError: "",
		},
		{
			name:          "Unsupported",
			provider:      "unsupported",
			apiKey:        "test",
			baseURL:       "",
			expectedError: "unsupported AI provider: unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SUBSTRATE_AI_PROVIDER", tt.provider)
			t.Setenv("SUBSTRATE_AI_API_KEY", tt.apiKey)
			t.Setenv("SUBSTRATE_AI_BASE_URL", tt.baseURL)

			client, err := NewAIClient()

			if tt.expectedError != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.expectedError)
				}
				if err.Error() != tt.expectedError {
					t.Errorf("expected error %q, got %q", tt.expectedError, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if client == nil {
					t.Fatal("expected non-nil client")
				}

				// Verify it implements AIClient
				var _ AIClient = client

				// Optional: we can't easily assert on the inner config without reflection or testing network calls,
				// but creating the client without panicking validates the factory.
			}
		})
	}
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}
