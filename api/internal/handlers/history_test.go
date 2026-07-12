package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/Substrate/api/internal/db"
	"github.com/google/uuid"
)

func TestHistoryHandler_Post(t *testing.T) {
	tests := []struct {
		name       string
		reqBody    RecordBreakingChangeRequest
		storeFunc  func(ctx context.Context, repoID uuid.UUID, orgName, repoName, gitSHA string, breakingChanges json.RawMessage) error
		wantStatus int
	}{
		{
			name: "success",
			reqBody: RecordBreakingChangeRequest{
				RepoID:          uuid.New(),
				OrgName:         "testorg",
				RepoName:        "testrepo",
				GitSHA:          "abcdef123456",
				BreakingChanges: json.RawMessage(`[{"type": "breaking"}]`),
			},
			storeFunc: func(ctx context.Context, repoID uuid.UUID, orgName, repoName, gitSHA string, breakingChanges json.RawMessage) error {
				return nil
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "missing required fields",
			reqBody: RecordBreakingChangeRequest{
				RepoID:  uuid.New(),
				OrgName: "testorg",
			},
			storeFunc:  nil,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &db.MockStore{
				RecordBreakingChangeFunc: tt.storeFunc,
			}

			handler := HistoryHandler(mockStore)

			body, _ := json.Marshal(tt.reqBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/history", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

func TestHistoryGetHandler(t *testing.T) {
	tests := []struct {
		name       string
		org        string
		repo       string
		limit      string
		storeFunc  func(ctx context.Context, orgName, repoName string, limit int) ([]db.BreakingChangeRecord, error)
		wantStatus int
		wantCount  int
	}{
		{
			name:  "success",
			org:   "testorg",
			repo:  "testrepo",
			limit: "5",
			storeFunc: func(ctx context.Context, orgName, repoName string, limit int) ([]db.BreakingChangeRecord, error) {
				return []db.BreakingChangeRecord{
					{
						ID:              uuid.New(),
						RepoID:          uuid.New(),
						OrgName:         orgName,
						RepoName:        repoName,
						GitSHA:          "abcdef123456",
						Timestamp:       time.Now(),
						BreakingChanges: json.RawMessage(`[{"type": "breaking"}]`),
					},
				}, nil
			},
			wantStatus: http.StatusOK,
			wantCount:  1,
		},
		{
			name:       "missing org",
			org:        "",
			repo:       "testrepo",
			limit:      "5",
			storeFunc:  nil,
			wantStatus: http.StatusBadRequest,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &db.MockStore{
				GetBreakingChangeHistoryFunc: tt.storeFunc,
			}

			handler := HistoryGetHandler(mockStore)

			url := "/api/v1/history/" + tt.org + "/" + tt.repo
			if tt.limit != "" {
				url += "?limit=" + tt.limit
			}
			req := httptest.NewRequest(http.MethodGet, url, nil)
			req.SetPathValue("org", tt.org)
			req.SetPathValue("repo", tt.repo)
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.wantStatus == http.StatusOK {
				var res []db.BreakingChangeRecord
				if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
					t.Fatalf("failed to parse response: %v", err)
				}
				if len(res) != tt.wantCount {
					t.Errorf("expected %d records, got %d", tt.wantCount, len(res))
				}
			}
		})
	}
}
