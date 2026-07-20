package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/postmortem"
	"github.com/stretchr/testify/assert"
)

func TestE2E_PostMortem(t *testing.T) {
	// Compile the Substrate binary
	cmd := exec.Command("go", "build", "-o", "substrate")
	err := cmd.Run()
	assert.NoError(t, err)
	defer os.Remove("./substrate")

	// Start mock Substrate API
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/graph/default" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{}"))
			return
		}

		assert.Equal(t, "/api/v1/changes", r.URL.Path)
		assert.NotEmpty(t, r.URL.Query().Get("since"))
		assert.NotEmpty(t, r.URL.Query().Get("until"))
		assert.Equal(t, "Bearer test-token", r.Header.Get("Authorization"))

		changes := []postmortem.BreakingChangeRecord{
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

	// Execute postmortem command
	postmortemCmd := exec.Command("./substrate", "postmortem", "--incident", "2023-10-10T12:00:00Z")
	postmortemCmd.Env = append(os.Environ(),
		"REGISTRY_API_TOKEN=test-token",
		"SUBSTRATE_API_URL="+mockAPI.URL,
		"SUBSTRATE_AI_PROVIDER=openai",
		"SUBSTRATE_AI_API_KEY=test-ai-key",
		"SUBSTRATE_AI_BASE_URL="+mockOpenAI.URL+"/v1",
		"SUBSTRATE_DISABLE_CACHE_SYNC=1",
	)

	out, err := postmortemCmd.CombinedOutput()
	assert.NoError(t, err, string(out))

	output := string(out)
	assert.Contains(t, output, "# Post-Mortem")
	assert.Contains(t, output, "Root cause: removed field id")
}
