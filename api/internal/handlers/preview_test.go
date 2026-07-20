package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestGetPreviewHandler(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		mockSetup  func(*db.MockStore)
		wantStatus int
	}{
		{
			name:  "Valid Token Not Expired",
			token: uuid.New().String(),
			mockSetup: func(m *db.MockStore) {
				m.GetPreviewSessionFunc = func(ctx context.Context, token uuid.UUID) (json.RawMessage, time.Time, error) {
					return json.RawMessage(`{"status":"ok"}`), time.Now().Add(1 * time.Hour), nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "Expired Token",
			token: uuid.New().String(),
			mockSetup: func(m *db.MockStore) {
				m.GetPreviewSessionFunc = func(ctx context.Context, token uuid.UUID) (json.RawMessage, time.Time, error) {
					return json.RawMessage(`{"status":"ok"}`), time.Now().Add(-1 * time.Hour), nil
				}
			},
			wantStatus: http.StatusGone,
		},
		{
			name:  "Invalid Token",
			token: "invalid-uuid",
			mockSetup: func(m *db.MockStore) {
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &db.MockStore{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockStore)
			}

			req := httptest.NewRequest("GET", "/api/v1/preview/"+tt.token, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("token", tt.token)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler := GetPreviewHandler(mockStore)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %v, got %v", tt.wantStatus, rr.Code)
			}
		})
	}
}

func TestExpirePreviewHandler(t *testing.T) {
	tests := []struct {
		name       string
		payload    interface{}
		mockSetup  func(*db.MockStore)
		wantStatus int
	}{
		{
			name: "Valid Request",
			payload: ExpirePreviewRequest{
				Org:      "testorg",
				Repo:     "testrepo",
				PRNumber: 1,
			},
			mockSetup: func(m *db.MockStore) {
				m.ExpirePreviewSessionsForPRFunc = func(ctx context.Context, prNumber int, orgName string, repoName string) error {
					return nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "Missing Fields",
			payload: ExpirePreviewRequest{
				Org: "testorg",
			},
			mockSetup:  func(m *db.MockStore) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := &db.MockStore{}
			if tt.mockSetup != nil {
				tt.mockSetup(mockStore)
			}

			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/api/v1/preview/expire-pr", bytes.NewBuffer(body))

			rr := httptest.NewRecorder()
			handler := ExpirePreviewHandler(mockStore)
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Errorf("expected status %v, got %v", tt.wantStatus, rr.Code)
			}
		})
	}
}
