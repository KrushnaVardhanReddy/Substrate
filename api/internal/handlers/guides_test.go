package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/go-chi/chi/v5"
)

type mockStoreForGuides struct {
	db.Store
	listRepoGuidesFunc func(ctx context.Context, org, repo string) ([]db.RepoGuide, error)
	getRepoGuideFunc   func(ctx context.Context, org, repo, slug string) (*db.RepoGuide, error)
}

func (m *mockStoreForGuides) ListRepoGuides(ctx context.Context, org, repo string) ([]db.RepoGuide, error) {
	return m.listRepoGuidesFunc(ctx, org, repo)
}

func (m *mockStoreForGuides) GetRepoGuide(ctx context.Context, org, repo, slug string) (*db.RepoGuide, error) {
	return m.getRepoGuideFunc(ctx, org, repo, slug)
}

func TestListGuidesHandler(t *testing.T) {
	mockStore := &mockStoreForGuides{
		listRepoGuidesFunc: func(ctx context.Context, org, repo string) ([]db.RepoGuide, error) {
			if org == "test-org" && repo == "test-repo" {
				return []db.RepoGuide{
					{FilePath: "guide1.md", Title: "Guide 1"},
				}, nil
			}
			return []db.RepoGuide{}, nil
		},
	}

	tests := []struct {
		name         string
		org          string
		repo         string
		expectedCode int
		expectedLen  int
	}{
		{"Valid Repo", "test-org", "test-repo", http.StatusOK, 1},
		{"Empty Repo", "other-org", "other-repo", http.StatusOK, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Get("/api/v1/docs/{org}/{repo}", ListGuidesHandler(mockStore))

			req := httptest.NewRequest("GET", "/api/v1/docs/"+tt.org+"/"+tt.repo, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}

			var result []db.RepoGuide
			json.NewDecoder(w.Body).Decode(&result)

			if len(result) != tt.expectedLen {
				t.Errorf("expected %d guides, got %d", tt.expectedLen, len(result))
			}
		})
	}
}

func TestGetGuideHandler(t *testing.T) {
	mockStore := &mockStoreForGuides{
		getRepoGuideFunc: func(ctx context.Context, org, repo, slug string) (*db.RepoGuide, error) {
			if slug == "guide1.md" {
				return &db.RepoGuide{FilePath: "guide1.md", Title: "Guide 1"}, nil
			}
			return nil, db.ErrNotFound
		},
	}

	tests := []struct {
		name         string
		slug         string
		expectedCode int
	}{
		{"Found", "guide1.md", http.StatusOK},
		{"Not Found", "missing.md", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := chi.NewRouter()
			r.Get("/api/v1/docs/{org}/{repo}/{slug}", GetGuideHandler(mockStore))
			r.Get("/api/v1/docs/{org}/{repo}/{slug}/*", GetGuideHandler(mockStore))

			req := httptest.NewRequest("GET", "/api/v1/docs/org/repo/"+tt.slug, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if w.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, w.Code)
			}
		})
	}
}
