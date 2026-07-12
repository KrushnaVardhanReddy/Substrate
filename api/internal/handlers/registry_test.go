package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/handlers"
)

func TestCanDeployHandler(t *testing.T) {
	tests := []struct {
		name                 string
		repoQuery            string
		commitQuery          string
		mockBreakingHistory  []db.BreakingChangeRecord
		mockEdges            []db.DependencyEdge
		breakingHistoryErr   error
		edgesErr             error
		expectedStatus       int
		expectedSafe         bool
		expectedBlockedCount int
	}{
		{
			name:           "missing repo",
			repoQuery:      "",
			commitQuery:    "a1b2c3d4",
			expectedStatus: http.StatusBadRequest,
			expectedSafe:   false,
		},
		{
			name:           "missing commit",
			repoQuery:      "owner/repo",
			commitQuery:    "",
			expectedStatus: http.StatusBadRequest,
			expectedSafe:   false,
		},
		{
			name:           "invalid repo format",
			repoQuery:      "owner_repo",
			commitQuery:    "a1b2c3d4",
			expectedStatus: http.StatusBadRequest,
			expectedSafe:   false,
		},
		{
			name:        "safe to deploy - no breaking change found",
			repoQuery:   "owner/repo",
			commitQuery: "a1b2c3d4",
			mockBreakingHistory: []db.BreakingChangeRecord{
				{GitSHA: "different_commit"},
			},
			expectedStatus: http.StatusOK,
			expectedSafe:   true,
		},
		{
			name:        "safe to deploy - breaking change, but no consumers",
			repoQuery:   "owner/repo",
			commitQuery: "a1b2c3d4",
			mockBreakingHistory: []db.BreakingChangeRecord{
				{GitSHA: "a1b2c3d4"},
			},
			mockEdges: []db.DependencyEdge{
				{ProviderFullName: "owner/other_repo", ConsumerFullName: "owner/consumer1"},
			},
			expectedStatus: http.StatusOK,
			expectedSafe:   true,
		},
		{
			name:        "deployment blocked - breaking change and consumers exist",
			repoQuery:   "owner/repo",
			commitQuery: "a1b2c3d4",
			mockBreakingHistory: []db.BreakingChangeRecord{
				{GitSHA: "a1b2c3d4"},
			},
			mockEdges: []db.DependencyEdge{
				{ProviderFullName: "owner/repo", ConsumerFullName: "owner/consumer1"},
				{ProviderFullName: "owner/repo", ConsumerFullName: "owner/consumer2"},
			},
			expectedStatus:       http.StatusConflict,
			expectedSafe:         false,
			expectedBlockedCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &db.MockStore{
				GetBreakingChangeHistoryFunc: func(ctx context.Context, orgName, repoName string, limit int) ([]db.BreakingChangeRecord, error) {
					return tt.mockBreakingHistory, tt.breakingHistoryErr
				},
				GetDependencyGraphFunc: func(ctx context.Context, orgName string) ([]db.DependencyEdge, error) {
					return tt.mockEdges, tt.edgesErr
				},
			}

			handler := handlers.CanDeployHandler(store)

			req := httptest.NewRequest("GET", "/api/v1/registry/can-deploy?repo="+tt.repoQuery+"&commit="+tt.commitQuery, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			var resp handlers.CanDeployResponse
			if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}
			if resp.Safe != tt.expectedSafe {
				t.Errorf("handler returned wrong safe status: got %v want %v", resp.Safe, tt.expectedSafe)
			}
			if len(resp.BlockedBy) != tt.expectedBlockedCount {
				t.Errorf("expected %d blocked consumers, got %d", tt.expectedBlockedCount, len(resp.BlockedBy))
			}
		})
	}
}
