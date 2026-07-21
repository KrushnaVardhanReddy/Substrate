package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/engine/pkg/ai"
)

func TestAIImpactHandler(t *testing.T) {
	// Unset required AI env vars to trigger mock mode in ai.NewAIClient
	os.Setenv("SUBSTRATE_AI_PROVIDER", "unsupported-to-force-error")

	handler := AIImpactHandler()

	reqBody := AIImpactRequest{
		Changes: []ai.Change{
			{RuleID: "TEST", Path: "/test", Description: "removed"},
		},
		Consumers: []ai.Repo{
			{Name: "test-repo"},
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/ai/impact", bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp AIImpactResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Explanation == "" {
		t.Errorf("expected non-empty explanation")
	}
}
