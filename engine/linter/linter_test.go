package linter

import (
	"context"
	"encoding/json"
	"github.com/spf13/viper"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestAnalyze(t *testing.T) {
	mockResponse := `
{
  "score": 85,
  "issues": [
    {
      "rule": "TOO_MANY_PARAMETERS",
      "path": "/users GET",
      "description": "Endpoint has 12 parameters, which is >10."
    }
  ]
}
`

	// Setup mock OpenAI server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")

		respBytes, _ := json.Marshal(map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"delta": map[string]interface{}{
						"content": mockResponse,
					},
				},
			},
		})

		w.Write([]byte("data: " + string(respBytes) + "\n\n"))
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer ts.Close()

	os.Setenv("SUBSTRATE_AI_PROVIDER", "openai")
	os.Setenv("SUBSTRATE_AI_API_KEY", "test-key")
	os.Setenv("SUBSTRATE_AI_BASE_URL", ts.URL)
	defer os.Unsetenv("SUBSTRATE_AI_PROVIDER")
	defer os.Unsetenv("SUBSTRATE_AI_API_KEY")
	defer os.Unsetenv("SUBSTRATE_AI_BASE_URL")

	schema := []byte("openapi: 3.0.0\ninfo:\n  title: Mock\n")

	score, issues, err := Analyze(context.Background(), schema)
	if err != nil {
		t.Fatalf("Analyze failed: %v", err)
	}

	if score != 85 {
		t.Errorf("Expected score 85, got %d", score)
	}

	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}

	if issues[0].Rule != "TOO_MANY_PARAMETERS" {
		t.Errorf("Expected rule TOO_MANY_PARAMETERS, got %s", issues[0].Rule)
	}
}

func TestAnalyzeFallback(t *testing.T) {
	viper.Reset()
	viper.AutomaticEnv()
	os.Unsetenv("SUBSTRATE_AI_BASE_URL")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("SUBSTRATE_AI_API_KEY")
	os.Unsetenv("SUBSTRATE_AI_BASE_URL")
	os.Unsetenv("OPENAI_API_KEY")
	os.Unsetenv("SUBSTRATE_AI_API_KEY")

	schema := []byte("openapi: 3.0.0\ninfo:\n  title: Mock\n")

	score, issues, err := Analyze(context.Background(), schema)
	if err != nil {
		t.Fatalf("Analyze fallback failed: %v", err)
	}

	if score != 90 {
		t.Errorf("Expected score 90, got %d", score)
	}

	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}

	if issues[0].Rule != "FALLBACK_MODE" {
		t.Errorf("Expected rule FALLBACK_MODE, got %s", issues[0].Rule)
	}
}

func init() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
}
