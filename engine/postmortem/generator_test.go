package postmortem

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePostMortem(t *testing.T) {
	// Start mock Substrate API
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/changes", r.URL.Path)
		assert.NotEmpty(t, r.URL.Query().Get("since"))
		assert.NotEmpty(t, r.URL.Query().Get("until"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		changes := []BreakingChangeRecord{
			{
				OrgName:         "testorg",
				RepoName:        "testrepo",
				GitSHA:          "abcdef",
				Timestamp:       time.Now().Add(-2 * time.Hour),
				BreakingChanges: json.RawMessage(`[{"description": "removed field id"}]`),
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(changes)
	}))
	defer mockAPI.Close()

	// Start mock OpenAI API
	mockOpenAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		w.Header().Set("Content-Type", "text/event-stream")
		w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"# Post-Mortem\\n\\n\"}}]}\n\n"))
		w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"Root cause: removed field id\"}}]}\n\n"))
		w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer mockOpenAI.Close()

	// Setup env vars
	os.Setenv("SUBSTRATE_AI_PROVIDER", "openai")
	os.Setenv("SUBSTRATE_AI_API_KEY", "test-ai-key")
	os.Setenv("SUBSTRATE_AI_BASE_URL", mockOpenAI.URL+"/v1")
	defer func() {
		os.Unsetenv("SUBSTRATE_AI_PROVIDER")
		os.Unsetenv("SUBSTRATE_AI_API_KEY")
		os.Unsetenv("SUBSTRATE_AI_BASE_URL")
	}()

	var buf bytes.Buffer
	err := GeneratePostMortem(context.Background(), mockAPI.URL, "test-token", time.Now(), &buf)

	assert.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "# Post-Mortem")
	assert.Contains(t, output, "Root cause: removed field id")
}

func TestGeneratePostMortem_NoChanges(t *testing.T) {
	// Start mock Substrate API returning empty changes
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]BreakingChangeRecord{})
	}))
	defer mockAPI.Close()

	var buf bytes.Buffer
	err := GeneratePostMortem(context.Background(), mockAPI.URL, "test-token", time.Now(), &buf)

	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "No breaking changes found")
}
