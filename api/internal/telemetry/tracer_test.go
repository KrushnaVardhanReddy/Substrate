package telemetry

import (
	"context"
	"os"
	"testing"
)

func TestInitTracerDisabled(t *testing.T) {
	os.Unsetenv("SUBSTRATE_OTEL_ENDPOINT")
	tp, err := InitTracer(context.Background(), "test-service")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tp != nil {
		t.Fatalf("expected nil TracerProvider, got %v", tp)
	}
}

func TestInitTracerEnabled(t *testing.T) {
	os.Setenv("SUBSTRATE_OTEL_ENDPOINT", "localhost:4318")
	defer os.Unsetenv("SUBSTRATE_OTEL_ENDPOINT")

	tp, err := InitTracer(context.Background(), "test-service")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tp == nil {
		t.Fatalf("expected TracerProvider to be non-nil")
	}
}
