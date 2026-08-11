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
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
)

type mockAIClient struct{}

func (m *mockAIClient) EmbedText(ctx context.Context, text string) ([]float64, error) {
	return []float64{0.1, 0.2, 0.3}, nil
}

func TestSchemaPruneHandler(t *testing.T) {
	mockStore := &db.MockStore{
		GetPrunedSchemaCacheFunc: func(ctx context.Context, arg sqlcgen.GetPrunedSchemaCacheParams) (sqlcgen.SchemaPruneCache, error) {
			return sqlcgen.SchemaPruneCache{}, context.DeadlineExceeded
		},
		GetSchemaFunc: func(ctx context.Context, org, repo string) (string, error) {
			if org == "testorg" && repo == "testrepo" {
				return `{"openapi":"3.0.0","info":{"title":"Test","version":"1"},"paths":{"/test":{"get":{"summary":"test"}}}}`, nil
			}
			return "", db.ErrNotFound
		},
		UpsertPrunedSchemaCacheFunc: func(ctx context.Context, arg sqlcgen.UpsertPrunedSchemaCacheParams) error {
			return nil
		},
	}

	mockAI := &mockAIClient{}
	handler := SchemaPruneHandler(mockStore, mockAI)

	tests := []struct {
		name           string
		body           map[string]string
		query          string
		expectedStatus int
	}{
		{
			name: "valid request",
			body: map[string]string{
				"org":    "testorg",
				"repo":   "testrepo",
				"intent": "test intent",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid max_endpoints",
			body: map[string]string{
				"org":    "testorg",
				"repo":   "testrepo",
				"intent": "test intent",
			},
			query:          "?max_endpoints=30",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing fields",
			body: map[string]string{
				"org": "testorg",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "schema not found",
			body: map[string]string{
				"org":    "notfound",
				"repo":   "repo",
				"intent": "test",
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest("POST", "/api/v1/schema/prune"+tt.query, bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestSchemaPruneHandler_CacheHit(t *testing.T) {
	mockStore := &db.MockStore{
		GetPrunedSchemaCacheFunc: func(ctx context.Context, arg sqlcgen.GetPrunedSchemaCacheParams) (sqlcgen.SchemaPruneCache, error) {
			return sqlcgen.SchemaPruneCache{
				PrunedSchema: []byte(`{"cached":"true"}`),
			}, nil
		},
	}

	mockAI := &mockAIClient{}
	handler := SchemaPruneHandler(mockStore, mockAI)

	body, _ := json.Marshal(map[string]string{
		"org":    "testorg",
		"repo":   "testrepo",
		"intent": "test intent",
	})
	req := httptest.NewRequest("POST", "/api/v1/schema/prune", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Body.String() != `{"cached":"true"}` {
		t.Errorf("expected cached response")
	}
}
