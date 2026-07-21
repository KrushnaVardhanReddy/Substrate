package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestGovernanceRulesHandler_ListRules(t *testing.T) {
	orgID := uuid.New()
	ruleID := uuid.New()

	tests := []struct {
		name           string
		orgName        string
		mockGetOrgID   func(ctx context.Context, name string) (uuid.UUID, error)
		mockGetRules   func(ctx context.Context, id uuid.UUID) ([]db.GovernanceRule, error)
		expectedStatus int
	}{
		{
			name:    "Success",
			orgName: "testorg",
			mockGetOrgID: func(ctx context.Context, name string) (uuid.UUID, error) {
				return orgID, nil
			},
			mockGetRules: func(ctx context.Context, id uuid.UUID) ([]db.GovernanceRule, error) {
				return []db.GovernanceRule{
					{ID: ruleID, OrgID: orgID, RuleText: "test rule", CreatedAt: time.Now()},
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "Org Not Found",
			orgName: "testorg",
			mockGetOrgID: func(ctx context.Context, name string) (uuid.UUID, error) {
				return uuid.Nil, errors.New("not found")
			},
			mockGetRules: func(ctx context.Context, id uuid.UUID) ([]db.GovernanceRule, error) {
				return nil, nil
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &db.MockStore{
				GetOrgIDByNameFunc:          tt.mockGetOrgID,
				GetGovernanceRulesByOrgFunc: tt.mockGetRules,
			}
			h := NewGovernanceRulesHandler(store)

			r := chi.NewRouter()
			r.Get("/orgs/{orgName}/rules", h.ListRules)

			req := httptest.NewRequest(http.MethodGet, "/orgs/"+tt.orgName+"/rules", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGovernanceRulesHandler_CreateRule(t *testing.T) {
	orgID := uuid.New()

	tests := []struct {
		name           string
		orgName        string
		body           GovernanceRuleRequest
		mockGetOrgID   func(ctx context.Context, name string) (uuid.UUID, error)
		mockUpsert     func(ctx context.Context, id uuid.UUID, text string) (uuid.UUID, error)
		expectedStatus int
	}{
		{
			name:    "Success",
			orgName: "testorg",
			body:    GovernanceRuleRequest{RuleText: "test rule"},
			mockGetOrgID: func(ctx context.Context, name string) (uuid.UUID, error) {
				return orgID, nil
			},
			mockUpsert: func(ctx context.Context, id uuid.UUID, text string) (uuid.UUID, error) {
				return uuid.New(), nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:    "Invalid Body",
			orgName: "testorg",
			body:    GovernanceRuleRequest{RuleText: ""},
			mockGetOrgID: func(ctx context.Context, name string) (uuid.UUID, error) {
				return orgID, nil
			},
			mockUpsert: func(ctx context.Context, id uuid.UUID, text string) (uuid.UUID, error) {
				return uuid.Nil, nil
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &db.MockStore{
				GetOrgIDByNameFunc:       tt.mockGetOrgID,
				UpsertGovernanceRuleFunc: tt.mockUpsert,
			}
			h := NewGovernanceRulesHandler(store)

			r := chi.NewRouter()
			r.Post("/orgs/{orgName}/rules", h.CreateRule)

			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/orgs/"+tt.orgName+"/rules", bytes.NewReader(body))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGovernanceRulesHandler_DeleteRule(t *testing.T) {
	orgID := uuid.New()
	ruleID := uuid.New()

	tests := []struct {
		name           string
		orgName        string
		ruleIDStr      string
		mockGetOrgID   func(ctx context.Context, name string) (uuid.UUID, error)
		mockDelete     func(ctx context.Context, rid uuid.UUID, oid uuid.UUID) error
		expectedStatus int
	}{
		{
			name:      "Success",
			orgName:   "testorg",
			ruleIDStr: ruleID.String(),
			mockGetOrgID: func(ctx context.Context, name string) (uuid.UUID, error) {
				return orgID, nil
			},
			mockDelete: func(ctx context.Context, rid uuid.UUID, oid uuid.UUID) error {
				return nil
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:      "Invalid Rule ID",
			orgName:   "testorg",
			ruleIDStr: "invalid-uuid",
			mockGetOrgID: func(ctx context.Context, name string) (uuid.UUID, error) {
				return orgID, nil
			},
			mockDelete: func(ctx context.Context, rid uuid.UUID, oid uuid.UUID) error {
				return nil
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &db.MockStore{
				GetOrgIDByNameFunc:       tt.mockGetOrgID,
				DeleteGovernanceRuleFunc: tt.mockDelete,
			}
			h := NewGovernanceRulesHandler(store)

			r := chi.NewRouter()
			r.Delete("/orgs/{orgName}/rules/{ruleID}", h.DeleteRule)

			req := httptest.NewRequest(http.MethodDelete, "/orgs/"+tt.orgName+"/rules/"+tt.ruleIDStr, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
