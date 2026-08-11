// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package server

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

func TestTierLimitsMiddleware(t *testing.T) {
	tests := []struct {
		name                 string
		method               string
		body                 string
		repoCount            int
		depCount             int
		existingRepos        []db.Repository
		expectedStatus       int
		expectedBodyContains string
	}{
		{
			name:           "GET requests pass through",
			method:         http.MethodGet,
			body:           "",
			repoCount:      5, // would fail if POST
			expectedStatus: http.StatusOK,
		},
		{
			name:           "POST free tier limits not reached",
			method:         http.MethodPost,
			body:           `{"org": "test-org", "consumer_repo": "org/repo1"}`,
			repoCount:      2,
			depCount:       4,
			expectedStatus: http.StatusOK,
		},
		{
			name:                 "POST repo limit exceeded",
			method:               http.MethodPost,
			body:                 `{"org": "test-org", "consumer_repo": "org/repo4"}`,
			repoCount:            3,
			expectedStatus:       http.StatusPaymentRequired,
			expectedBodyContains: "Free tier limit reached. Please upgrade to Pro.",
		},
		{
			name:                 "POST repo limit exceeded with dependencies",
			method:               http.MethodPost,
			body:                 `{"org": "test-org", "dependencies": [{"provider_repo": "org/provider4"}]}`,
			repoCount:            3,
			expectedStatus:       http.StatusPaymentRequired,
			expectedBodyContains: "Free tier limit reached. Please upgrade to Pro.",
		},
		{
			name:                 "POST downstream dependency limit exceeded",
			method:               http.MethodPost,
			body:                 `{"org": "test-org", "consumer_repo": "org/repo1", "dependencies": [{"provider_repo": "org/provider1"}]}`,
			repoCount:            1,
			depCount:             5,
			expectedStatus:       http.StatusPaymentRequired,
			expectedBodyContains: "Free tier limit reached. Please upgrade to Pro.",
		},
		{
			name:           "POST downstream dependency existing skip limit",
			method:         http.MethodPost,
			body:           `{"org": "test-org", "consumer_repo": "org/repo1", "dependencies": [{"provider_repo": "org/provider1"}]}`,
			repoCount:      1,
			depCount:       5,
			existingRepos:  []db.Repository{{FullName: "org/repo1"}},
			expectedStatus: http.StatusPaymentRequired, // It's still a new dependency if we don't mock GetConsumers. (mock returns none by default)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &db.MockStore{
				CountReposByOrgFunc: func(ctx context.Context, orgName string) (int, error) {
					return tt.repoCount, nil
				},
				ListReposByOrgFunc: func(ctx context.Context, orgName string) ([]db.Repository, error) {
					return tt.existingRepos, nil
				},
				CountDownstreamDependenciesFunc: func(ctx context.Context, providerFullName string) (int, error) {
					return tt.depCount, nil
				},
			}

			mw := TierLimitsMiddleware(store)
			handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(tt.method, "/test", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}

			if tt.expectedBodyContains != "" {
				if !bytes.Contains(rr.Body.Bytes(), []byte(tt.expectedBodyContains)) {
					t.Errorf("expected body to contain %q, got %q", tt.expectedBodyContains, rr.Body.String())
				}
			}
		})
	}
}
