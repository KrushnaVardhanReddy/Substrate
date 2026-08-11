// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestServeDashboard(t *testing.T) {
	r := chi.NewRouter()
	ServeDashboard(r)

	// Since we mock or include static/index.html via the filesystem/embed,
	// test standard route fallback behavior

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		checkBody      func(body string) bool
	}{
		{
			name:           "API route should 404 from Catch-All",
			path:           "/api/v1/unknown",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Root path should serve index.html (SPA)",
			path:           "/",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Unknown non-API path should fallback to index.html",
			path:           "/dashboard/settings",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()

			if res.StatusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, res.StatusCode)
			}

			if tt.checkBody != nil {
				body, _ := io.ReadAll(res.Body)
				if !tt.checkBody(string(body)) {
					t.Errorf("body content did not match expectations")
				}
			}
		})
	}
}
