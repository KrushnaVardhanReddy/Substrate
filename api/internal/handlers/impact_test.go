package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

func TestImpactHandler(t *testing.T) {
	mockStore := &db.MockStore{
		GetDependencyGraphFunc: func(ctx context.Context, orgName string) ([]db.DependencyEdge, error) {
			return []db.DependencyEdge{
				{ProviderFullName: "myorg/repo", ConsumerFullName: "myorg/consumer-a"},
				{ProviderFullName: "myorg/repo", ConsumerFullName: "myorg/consumer-b"},
				{ProviderFullName: "myorg/consumer-a", ConsumerFullName: "myorg/consumer-c"},
				{ProviderFullName: "myorg/other", ConsumerFullName: "myorg/consumer-d"},
			}, nil
		},
	}

	tests := []struct {
		name           string
		org            string
		repo           string
		expectedStatus int
		expectedResp   ImpactResponse
	}{
		{
			name:           "valid impact calculation",
			org:            "myorg",
			repo:           "repo",
			expectedStatus: http.StatusOK,
			expectedResp: ImpactResponse{
				Provider:      "myorg/repo",
				RiskScore:     3,
				ImpactedRepos: []string{"myorg/consumer-a", "myorg/consumer-b", "myorg/consumer-c"},
			},
		},
		{
			name:           "missing repo parameter",
			org:            "myorg",
			repo:           "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing org parameter",
			org:            "",
			repo:           "repo",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/api/v1/impact/"+tt.org+"/"+tt.repo, nil)
			if err != nil {
				t.Fatal(err)
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("org", tt.org)
			rctx.URLParams.Add("repo", tt.repo)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler := ImpactHandler(mockStore)

			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK {
				var resp ImpactResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
					t.Fatal(err)
				}

				if resp.Provider != tt.expectedResp.Provider {
					t.Errorf("expected provider '%s', got '%s'", tt.expectedResp.Provider, resp.Provider)
				}

				if resp.RiskScore != tt.expectedResp.RiskScore {
					t.Errorf("expected risk score %d, got %d", tt.expectedResp.RiskScore, resp.RiskScore)
				}

				if !reflect.DeepEqual(resp.ImpactedRepos, tt.expectedResp.ImpactedRepos) {
					t.Errorf("expected impacted repos %v, got %v", tt.expectedResp.ImpactedRepos, resp.ImpactedRepos)
				}
			}
		})
	}
}
