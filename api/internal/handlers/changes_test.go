// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestChangesHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		url            string
		expectedStatus int
		mockStore      *db.MockStore
	}{
		{
			name:           "Method Not Allowed",
			method:         http.MethodPost,
			url:            "/api/v1/changes",
			expectedStatus: http.StatusMethodNotAllowed,
			mockStore:      &db.MockStore{},
		},
		{
			name:           "Missing Since",
			method:         http.MethodGet,
			url:            "/api/v1/changes",
			expectedStatus: http.StatusBadRequest,
			mockStore:      &db.MockStore{},
		},
		{
			name:           "Invalid Since",
			method:         http.MethodGet,
			url:            "/api/v1/changes?since=invalid",
			expectedStatus: http.StatusBadRequest,
			mockStore:      &db.MockStore{},
		},
		{
			name:           "Invalid Until",
			method:         http.MethodGet,
			url:            "/api/v1/changes?since=2023-10-10T12:00:00Z&until=invalid",
			expectedStatus: http.StatusBadRequest,
			mockStore:      &db.MockStore{},
		},
		{
			name:           "Success",
			method:         http.MethodGet,
			url:            "/api/v1/changes?since=2023-10-10T12:00:00Z&until=2023-10-11T12:00:00Z",
			expectedStatus: http.StatusOK,
			mockStore: &db.MockStore{
				GetBreakingChangesBetweenFunc: func(ctx context.Context, since, until time.Time) ([]db.BreakingChangeRecord, error) {
					return []db.BreakingChangeRecord{
						{OrgName: "org1", RepoName: "repo1", GitSHA: "sha1", BreakingChanges: json.RawMessage(`[]`)},
					}, nil
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.url, nil)
			rr := httptest.NewRecorder()

			handler := ChangesHandler(tt.mockStore)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
