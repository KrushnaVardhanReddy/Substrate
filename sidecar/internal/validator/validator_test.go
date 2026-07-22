package validator

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/sidecar/internal/proxy"
)

func TestValidator(t *testing.T) {
	mockSchema := `
openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      responses:
        '200':
          description: OK
`

	// Mock Substrate Registry server
	registryServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/v1/schema/") {
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockSchema))
	}))
	defer registryServer.Close()

	failureChan := make(chan ValidationFailure, 10)

	v := New(registryServer.URL, "test-token", "test-org", "test-repo", 1*time.Minute, failureChan)

	// Fetch schema synchronously for testing
	ctx := context.Background()
	v.fetchSchema(ctx)

	if v.schema == nil {
		t.Fatal("expected schema to be loaded")
	}

	// Test a valid request
	validReq := proxy.SampledRequest{
		Method: http.MethodGet,
		Path:   "/users",
		Headers: http.Header{},
	}

	v.ValidateRequest(ctx, validReq)

	select {
	case <-failureChan:
		t.Error("expected no validation failure for valid request")
	default:
		// Passed
	}

	// Test an invalid request (endpoint not found)
	invalidReq := proxy.SampledRequest{
		Method: http.MethodGet,
		Path:   "/hidden/endpoint",
		Headers: http.Header{},
	}

	v.ValidateRequest(ctx, invalidReq)

	select {
	case failure := <-failureChan:
		if failure.Method != "GET" || failure.Path != "/hidden/endpoint" {
			t.Errorf("unexpected failure details: %+v", failure)
		}
		if !strings.Contains(failure.ErrorMessage, "endpoint not found") {
			t.Errorf("expected 'endpoint not found' error, got: %s", failure.ErrorMessage)
		}
	default:
		t.Error("expected validation failure for invalid request")
	}
}
