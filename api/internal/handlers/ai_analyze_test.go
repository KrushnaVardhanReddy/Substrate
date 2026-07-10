package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAIAnalyzeHandler_Mock(t *testing.T) {
	// Substrate_AI_BASE_URL is purposely not set to test the mock response
	t.Setenv("SUBSTRATE_AI_BASE_URL", "")

	reqBody := `{"org":"test","current_schema":"type A{}","proposed_schema":"type B{}","schema_type":"graphql"}`
	req, err := http.NewRequest("POST", "/api/v1/ai/analyze", bytes.NewBufferString(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := AIAnalyzeHandler()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, "text/event-stream")
	}

	bodyStr := rr.Body.String()

	if !strings.Contains(bodyStr, `"type":"thinking"`) {
		t.Errorf("response body does not contain thinking event")
	}

	if !strings.Contains(bodyStr, `"type":"finding"`) {
		t.Errorf("response body does not contain finding event")
	}

	if !strings.Contains(bodyStr, `"type":"done"`) {
		t.Errorf("response body does not contain done event")
	}
}

func TestAIAnalyzeHandler_InvalidJSON(t *testing.T) {
	reqBody := `not valid json`
	req, err := http.NewRequest("POST", "/api/v1/ai/analyze", bytes.NewBufferString(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := AIAnalyzeHandler()

	handler.ServeHTTP(rr, req)

	// Since SSE headers are set first, we still get a 200 OK
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	contentType := rr.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		t.Errorf("handler returned wrong content type: got %v want %v", contentType, "text/event-stream")
	}

	bodyStr := rr.Body.String()

	if !strings.Contains(bodyStr, `"type":"error"`) {
		t.Errorf("response body does not contain error event")
	}

	if !strings.Contains(bodyStr, `"type":"done"`) {
		t.Errorf("response body does not contain done event")
	}
}
