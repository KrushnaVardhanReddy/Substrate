package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/workers"
	"github.com/google/uuid"
)

func TestSaveDiffHandler(t *testing.T) {
	store := &db.MockStore{}
	mockEnqueuer := &workers.MockJobEnqueuer{}
	handler := SaveDiffHandler(store, mockEnqueuer)

	tests := []struct {
		name           string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Valid request",
			body: map[string]interface{}{
				"diff_report": map[string]interface{}{
					"status": "diff found",
				},
				"org": "test-org",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Empty body",
			body:           nil,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing diff_report",
			body:           map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			}
			req := httptest.NewRequest("POST", "/api/v1/diff", bytes.NewBuffer(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetDiffHandler(t *testing.T) {
	store := &db.MockStore{}
	handler := GetDiffHandler(store)

	tests := []struct {
		name           string
		id             string
		expectedStatus int
	}{
		{
			name:           "Valid UUID",
			id:             uuid.New().String(),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid UUID",
			id:             "invalid-uuid",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/api/v1/diff/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
