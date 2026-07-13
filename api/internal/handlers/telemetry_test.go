package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

func TestTelemetryHandler(t *testing.T) {
	var wg sync.WaitGroup

	mockStore := &db.MockStore{
		UpdateDependencyConfidenceFunc: func(ctx context.Context, consumerFullName, providerURL string, boostAmount float64) error {
			defer wg.Done()
			if consumerFullName == "fail" {
				return fmt.Errorf("mock error")
			}
			return nil
		},
	}

	tests := []struct {
		name       string
		payload    string
		wantStatus int
		expectSync bool
	}{
		{
			name:       "valid payload",
			payload:    `[{"service.name":"frontend","http.url":"https://api.backend.com/users"}]`,
			wantStatus: http.StatusAccepted,
			expectSync: true,
		},
		{
			name:       "invalid payload",
			payload:    `{"invalid":"json"}`,
			wantStatus: http.StatusBadRequest,
			expectSync: false,
		},
		{
			name:       "process failure is async but returns accepted",
			payload:    `[{"service.name":"fail","http.url":"https://api.backend.com/users"}]`,
			wantStatus: http.StatusAccepted,
			expectSync: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectSync {
				wg.Add(1)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/telemetry/traces", bytes.NewBufferString(tt.payload))
			w := httptest.NewRecorder()

			handler := TelemetryHandler(mockStore)
			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.expectSync {
				wg.Wait()
			}
		})
	}
}
