package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

func TestArgoDriftWebhook(t *testing.T) {
	tests := []struct {
		name           string
		payload        ArgoWebhookPayload
		mockAnomalies  []db.DriftAnomaly
		mockErr        error
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "No anomalies, should return OK",
			payload: ArgoWebhookPayload{
				OrgName:  "test-org",
				RepoName: "test-repo",
			},
			mockAnomalies:  []db.DriftAnomaly{},
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
		},
		{
			name: "Severe anomalies, should return Not Acceptable",
			payload: ArgoWebhookPayload{
				OrgName:  "test-org",
				RepoName: "test-repo",
			},
			mockAnomalies: []db.DriftAnomaly{
				{
					OrgName:      "test-org",
					RepoName:     "test-repo",
					Method:       "GET",
					Path:         "/api/v1/test",
					ErrorMessage: "Schema violation",
				},
			},
			expectedStatus: http.StatusNotAcceptable,
			expectedBody:   "Severe schema violations detected\n",
		},
		{
			name: "Missing org or repo",
			payload: ArgoWebhookPayload{
				OrgName: "test-org",
			},
			mockAnomalies:  []db.DriftAnomaly{},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Missing org_name or repo_name\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &db.MockStore{
				GetDriftAnomaliesFunc: func(ctx context.Context, orgName, repoName string) ([]db.DriftAnomaly, error) {
					return tt.mockAnomalies, tt.mockErr
				},
			}

			handler := ArgoDriftWebhookHandler(store)

			body, _ := json.Marshal(tt.payload)
			req, err := http.NewRequest("POST", "/api/v1/argo/drift-webhook", bytes.NewBuffer(body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, rr.Body.String())
			}
		})
	}
}
