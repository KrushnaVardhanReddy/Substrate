// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"context"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/workers"
	"github.com/google/uuid"
)

func TestSaveDiffHandler(t *testing.T) {
	store := &db.MockStore{}
	mockEnqueuer := &workers.MockJobEnqueuer{}
	mockClient := &github.MockClient{
		GetFileContentFunc: func(ctx context.Context, owner, repo, path string) (string, error) {
			if path == ".substrate/SCHEMAOWNERS.yaml" {
				return "- path: \"/*\"\n  reviewers: [\"@myorg/api-platform\"]\n", nil
			}
			return "", nil
		},
		RequestReviewersFunc: func(ctx context.Context, owner, repo string, pullNumber int, reviewers []string) error {
			return nil
		},
	}
	handler := SaveDiffHandler(store, mockEnqueuer, mockClient)

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
