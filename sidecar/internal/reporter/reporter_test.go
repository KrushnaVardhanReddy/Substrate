package reporter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/sidecar/internal/validator"
)

func TestReporter(t *testing.T) {
	var receivedPayload DriftAnomalyPayload

	// Mock Substrate Telemetry server
	telemetryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/telemetry/drift" {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}

		if err := json.NewDecoder(r.Body).Decode(&receivedPayload); err != nil {
			t.Fatalf("failed to decode payload: %v", err)
		}

		w.WriteHeader(http.StatusAccepted)
	}))
	defer telemetryServer.Close()

	r := New(telemetryServer.URL, "test-token", "test-org", "test-repo")

	failure := validator.ValidationFailure{
		Method:       "GET",
		Path:         "/hidden",
		ErrorMessage: "not found",
	}

	r.ReportAnomaly(context.Background(), failure)

	if receivedPayload.OrgName != "test-org" {
		t.Errorf("expected org test-org, got %s", receivedPayload.OrgName)
	}
	if receivedPayload.Method != "GET" {
		t.Errorf("expected method GET, got %s", receivedPayload.Method)
	}
	if receivedPayload.Path != "/hidden" {
		t.Errorf("expected path /hidden, got %s", receivedPayload.Path)
	}
}
