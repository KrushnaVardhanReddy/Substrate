package discovery

import (
	"context"
	"fmt"
	"testing"

	"github.com/KrushnaVardhanReddy/Substrate/api/internal/db"
)

func TestExtractBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"http basic", "http://example.com/api/v1/users", "http://example.com"},
		{"https basic", "https://api.github.com/repos/test", "https://api.github.com"},
		{"no path", "https://example.com", "https://example.com"},
		{"port included", "http://localhost:8080/test", "http://localhost:8080"},
		{"invalid format", "invalid-url", "invalid-url"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractBaseURL(tt.input); got != tt.expected {
				t.Errorf("ExtractBaseURL(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestProcessRuntimeSignals(t *testing.T) {
	mockStore := &db.MockStore{
		UpdateDependencyConfidenceFunc: func(ctx context.Context, consumerFullName, providerURL string, boostAmount float64) error {
			if consumerFullName == "fail-service" {
				return fmt.Errorf("mock error")
			}
			if consumerFullName != "frontend-service" {
				t.Errorf("expected consumer frontend-service, got %s", consumerFullName)
			}
			if providerURL != "https://api.backend.com" {
				t.Errorf("expected provider https://api.backend.com, got %s", providerURL)
			}
			if boostAmount != 20.0 {
				t.Errorf("expected boost 20.0, got %f", boostAmount)
			}
			return nil
		},
	}

	t.Run("success", func(t *testing.T) {
		spans := []OTelSpan{
			{
				ServiceName: "frontend-service",
				HTTPURL:     "https://api.backend.com/users",
			},
		}

		err := ProcessRuntimeSignals(context.Background(), mockStore, spans)
		if err != nil {
			t.Errorf("ProcessRuntimeSignals failed: %v", err)
		}
	})

	t.Run("failure", func(t *testing.T) {
		spans := []OTelSpan{
			{
				ServiceName: "fail-service",
				HTTPURL:     "https://fail.com",
			},
		}

		err := ProcessRuntimeSignals(context.Background(), mockStore, spans)
		if err == nil {
			t.Errorf("ProcessRuntimeSignals expected error, got nil")
		}
	})
}
