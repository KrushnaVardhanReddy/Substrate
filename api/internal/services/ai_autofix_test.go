package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestAIAutofixHandler(t *testing.T) {
	// Ensure SUBSTRATE_AI_BASE_URL is not set for mock mode
	os.Unsetenv("SUBSTRATE_AI_BASE_URL")

	t.Run("returns mock autofix response when SUBSTRATE_AI_BASE_URL is not set", func(t *testing.T) {
		reqBody := AIAutofixRequest{
			ProviderRepo:   "test/repo",
			SchemaType:     "openapi",
			CurrentSchema:  "type A",
			ProposedSchema: "type B",
			BreakingChanges: []BreakingChange{
				{
					RuleID:      "test-rule",
					Path:        "test-path",
					Description: "test desc",
					Severity:    "BREAKING",
				},
			},
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req, err := http.NewRequest(http.MethodPost, "/api/v1/ai/autofix", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := AIAutofixHandler()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		var resp AIAutofixResponse
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}

		if resp.PatchLanguage != "yaml" {
			t.Errorf("expected PatchLanguage to be yaml, got %s", resp.PatchLanguage)
		}
	})

	t.Run("returns mock unified diff when ConsumerSourceCode is set", func(t *testing.T) {
		reqBody := AIAutofixRequest{
			ProviderRepo:   "test/repo",
			SchemaType:     "openapi",
			CurrentSchema:  "type A",
			ProposedSchema: "type B",
			BreakingChanges: []BreakingChange{
				{
					RuleID:      "test-rule",
					Path:        "test-path",
					Description: "test desc",
					Severity:    "BREAKING",
				},
			},
			ConsumerSourceCode: "func test() {}",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		req, err := http.NewRequest(http.MethodPost, "/api/v1/ai/autofix", bytes.NewReader(bodyBytes))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := AIAutofixHandler()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		if ctype := rr.Header().Get("Content-Type"); ctype != "application/json" {
			t.Errorf("handler returned wrong content type: got %v want %v", ctype, "application/json")
		}

		var resp AIAutofixResponse
		if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
			t.Fatal(err)
		}

		if !resp.MockMode {
			t.Error("expected MockMode to be true")
		}

		if resp.Explanation == "" {
			t.Error("expected non-empty Explanation")
		}

		if resp.SafePatch == "" {
			t.Error("expected non-empty SafePatch")
		}

		if resp.PatchLanguage != "diff" {
			t.Errorf("expected PatchLanguage to be diff, got %s", resp.PatchLanguage)
		}
	})

	t.Run("returns 400 on invalid JSON body", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/api/v1/ai/autofix", bytes.NewReader([]byte("not valid json")))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := AIAutofixHandler()

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		}
	})
}
