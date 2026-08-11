// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/go-chi/chi/v5"
)

func TestRegisterWebhookHandler(t *testing.T) {
	tests := []struct {
		name           string
		org            string
		body           map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Valid request",
			org:  "acme-corp",
			body: map[string]interface{}{
				"url":    "https://example.com/webhook",
				"secret": "my-secret",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Missing organization",
			org:            "",
			body:           map[string]interface{}{"url": "https://example.com"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing url",
			org:            "acme-corp",
			body:           map[string]interface{}{"secret": "secret"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty body",
			org:            "acme-corp",
			body:           nil,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &db.MockStore{
				RegisterWebhookFunc: func(ctx context.Context, config db.WebhookConfig) error {
					return nil
				},
			}
			handler := RegisterWebhookHandler(mockStore)

			var bodyBytes []byte
			if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest("POST", "/api/v1/org/"+tt.org+"/webhooks", bytes.NewBuffer(bodyBytes))

			// Inject chi URL param if org is present
			rctx := chi.NewRouteContext()
			if tt.org != "" {
				rctx.URLParams.Add("org", tt.org)
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
