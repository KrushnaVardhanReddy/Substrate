package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockSyncClient struct {
	syncErr error
}

func (m *mockSyncClient) SyncSchema(ctx context.Context, schema string) error {
	return m.syncErr
}

func TestHandlerWithClient(t *testing.T) {
	tests := []struct {
		name       string
		payload    WebhookPayload
		statusCode int
		invalidJSON bool
		syncErr    error
	}{
		{
			name: "valid safe payload",
			payload: WebhookPayload{
				Event:  "schema_update",
				Status: "SAFE",
				Type:   "openapi",
				Schema: "{}",
			},
			statusCode: http.StatusOK,
		},
		{
			name: "valid approved payload",
			payload: WebhookPayload{
				Event:  "schema_update",
				Status: "APPROVED",
				Type:   "openapi",
				Schema: "{}",
			},
			statusCode: http.StatusOK,
		},
		{
			name: "ignored status",
			payload: WebhookPayload{
				Event:  "schema_update",
				Status: "PENDING",
				Type:   "openapi",
				Schema: "{}",
			},
			statusCode: http.StatusOK,
		},
		{
			name: "ignored type",
			payload: WebhookPayload{
				Event:  "schema_update",
				Status: "SAFE",
				Type:   "graphql",
				Schema: "{}",
			},
			statusCode: http.StatusOK,
		},
		{
			name: "invalid json",
			invalidJSON: true,
			statusCode: http.StatusBadRequest,
		},
		{
			name: "sync error",
			payload: WebhookPayload{
				Event:  "schema_update",
				Status: "SAFE",
				Type:   "openapi",
				Schema: "{}",
			},
			statusCode: http.StatusInternalServerError,
			syncErr:    errors.New("sync failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.invalidJSON {
				body = []byte("invalid")
			} else {
				body, _ = json.Marshal(tt.payload)
			}
			req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
			w := httptest.NewRecorder()

			client := &mockSyncClient{syncErr: tt.syncErr}
			HandlerWithClient(client).ServeHTTP(w, req)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	handler := Handler()
	if handler == nil {
		t.Error("Expected Handler to return a non-nil http.HandlerFunc")
	}
}
