package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/go-chi/chi/v5"
)

func TestChangelogHandler(t *testing.T) {
	mockStore := &db.MockStore{
		GetDiffReportsByRepoFunc: func(ctx context.Context, orgName, repoName string, limit int) ([]db.DiffReportRecord, error) {
			reportData := `{
				"summary": {
					"breaking_count": 1,
					"safe_count": 1
				},
				"breaking_changes": [
					{"rule_id": "ENDPOINT_REMOVED"}
				],
				"safe_changes": [
					{"rule_id": "ENDPOINT_ADDED"}
				]
			}`

			createdAt, _ := time.Parse("2006-01-02T15:04:05Z", "2023-10-25T10:00:00Z")

			return []db.DiffReportRecord{
			    {
			        ReportData: []byte(reportData),
			        CreatedAt: createdAt,
			    },
			}, nil
		},
	}

	handler := ChangelogHandler(mockStore)

	req := httptest.NewRequest("GET", "/api/v1/changelog/test-org/test-repo", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("org", "test-org")
	rctx.URLParams.Add("repo", "test-repo")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var changelog []ChangelogEntry
	if err := json.NewDecoder(w.Body).Decode(&changelog); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(changelog) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(changelog))
	}

	entry := changelog[0]
	if entry.Date != "2023-10-25T10:00:00Z" {
		t.Errorf("expected date 2023-10-25T10:00:00Z, got %s", entry.Date)
	}
	if entry.EndpointsAdded != 1 {
		t.Errorf("expected 1 endpoint added, got %d", entry.EndpointsAdded)
	}
	if entry.EndpointsRemoved != 1 {
		t.Errorf("expected 1 endpoint removed, got %d", entry.EndpointsRemoved)
	}
}
