package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

func TestBadgesHandler(t *testing.T) {
	mockStore := &db.MockStore{
		ListReposByOrgFunc: func(ctx context.Context, orgName string) ([]db.Repository, error) {
			if orgName == "testorg" {
				return []db.Repository{
					{
						ID:   uuid.MustParse("00000000-0000-0000-0000-000000000001"),
						Name: "repo0",
					},
					{
						ID:   uuid.MustParse("00000000-0000-0000-0000-000000000002"),
						Name: "repo1",
					},
					{
						ID:   uuid.MustParse("00000000-0000-0000-0000-000000000003"),
						Name: "repo3",
					},
					{
						ID:   uuid.MustParse("00000000-0000-0000-0000-000000000004"),
						Name: "repo10",
					},
				}, nil
			}
			return nil, nil
		},
		CountRecentBreakingChangesFunc: func(ctx context.Context, repoID uuid.UUID, since time.Time) (int, error) {
			switch repoID.String() {
			case "00000000-0000-0000-0000-000000000001":
				return 0, nil
			case "00000000-0000-0000-0000-000000000002":
				return 1, nil
			case "00000000-0000-0000-0000-000000000003":
				return 3, nil
			case "00000000-0000-0000-0000-000000000004":
				return 10, nil
			}
			return 0, nil
		},
	}

	tests := []struct {
		name           string
		org            string
		repo           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "missing repo",
			org:            "testorg",
			repo:           "nonexistent",
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "A+ score",
			org:            "testorg",
			repo:           "repo0",
			expectedStatus: http.StatusOK,
			expectedBody:   "A+ | 0 breaks in 90 days",
		},
		{
			name:           "B score",
			org:            "testorg",
			repo:           "repo1",
			expectedStatus: http.StatusOK,
			expectedBody:   "B | 1 breaks in 90 days",
		},
		{
			name:           "C score",
			org:            "testorg",
			repo:           "repo3",
			expectedStatus: http.StatusOK,
			expectedBody:   "C | 3 breaks in 90 days",
		},
		{
			name:           "F score",
			org:            "testorg",
			repo:           "repo10",
			expectedStatus: http.StatusOK,
			expectedBody:   "F | 10 breaks in 90 days",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/api/badges/"+tc.org+"/"+tc.repo, nil)
			w := httptest.NewRecorder()

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("org", tc.org)
			rctx.URLParams.Add("repo", tc.repo)
			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			handler := BadgesHandler(mockStore)
			handler.ServeHTTP(w, r)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, w.Code)
			}

			if tc.expectedStatus == http.StatusOK {
				if w.Header().Get("Content-Type") != "image/svg+xml" {
					t.Errorf("expected Content-Type image/svg+xml, got %s", w.Header().Get("Content-Type"))
				}
				if w.Header().Get("Cache-Control") != "public, max-age=86400" {
					t.Errorf("expected Cache-Control public, max-age=86400, got %s", w.Header().Get("Cache-Control"))
				}

				if !strings.Contains(w.Body.String(), tc.expectedBody) {
					t.Errorf("expected body to contain %q", tc.expectedBody)
				}
			}
		})
	}
}
