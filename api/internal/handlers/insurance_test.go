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
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestInsuranceGetPolicyHandler(t *testing.T) {
	orgID := uuid.New()

	mockStore := &db.MockStore{
		GetOrgIDByNameFunc: func(ctx context.Context, name string) (uuid.UUID, error) {
			return orgID, nil
		},
		GetInsurancePolicyFunc: func(ctx context.Context, id uuid.UUID) (*db.InsurancePolicy, error) {
			if id == orgID {
				return &db.InsurancePolicy{
					OrgID:            orgID,
					PolicyLimitCents: 500000,
				}, nil
			}
			return nil, nil
		},
	}

	r := chi.NewRouter()
	r.Get("/api/v1/org/{org}/insurance/policy", InsuranceGetPolicyHandler(mockStore))

	req := httptest.NewRequest("GET", "/api/v1/org/"+orgID.String()+"/insurance/policy", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var policy db.InsurancePolicy
	err := json.NewDecoder(rr.Body).Decode(&policy)
	assert.NoError(t, err)
	assert.Equal(t, orgID, policy.OrgID)
	assert.Equal(t, int64(500000), policy.PolicyLimitCents)
}

func TestInsuranceGetClaimsHandler(t *testing.T) {
	orgID := uuid.New()
	policyID := uuid.New()

	mockStore := &db.MockStore{
		GetOrgIDByNameFunc: func(ctx context.Context, name string) (uuid.UUID, error) {
			return orgID, nil
		},
		GetInsuranceClaimsFunc: func(ctx context.Context, id uuid.UUID) ([]db.InsuranceClaim, error) {
			if id == orgID {
				return []db.InsuranceClaim{
					{
						OrgID:       orgID,
						PolicyID:    policyID,
						GithubPRUrl: "https://github.com/owner/repo/pull/1",
						Status:      "PENDING",
					},
				}, nil
			}
			return nil, nil
		},
	}

	r := chi.NewRouter()
	r.Get("/api/v1/org/{org}/insurance/claims", InsuranceGetClaimsHandler(mockStore))

	req := httptest.NewRequest("GET", "/api/v1/org/"+orgID.String()+"/insurance/claims", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var claims []db.InsuranceClaim
	err := json.NewDecoder(rr.Body).Decode(&claims)
	assert.NoError(t, err)
	assert.Len(t, claims, 1)
	assert.Equal(t, "https://github.com/owner/repo/pull/1", claims[0].GithubPRUrl)
}

func TestInsuranceFileClaimHandler(t *testing.T) {
	orgID := uuid.New()
	policyID := uuid.New()

	mockStore := &db.MockStore{
		GetOrgIDByNameFunc: func(ctx context.Context, name string) (uuid.UUID, error) {
			return orgID, nil
		},
		GetInsurancePolicyFunc: func(ctx context.Context, id uuid.UUID) (*db.InsurancePolicy, error) {
			if id == orgID {
				return &db.InsurancePolicy{
					ID:               policyID,
					OrgID:            orgID,
					PolicyLimitCents: 500000,
				}, nil
			}
			return nil, nil
		},
		CreateInsuranceClaimFunc: func(ctx context.Context, claim db.InsuranceClaim) (uuid.UUID, error) {
			return claim.ID, nil
		},
	}

	mockGHClient := &github.MockClient{
		ListCheckRunsForRefFunc: func(ctx context.Context, owner, repo, ref string) ([]github.CheckRun, error) {
			return []github.CheckRun{}, nil
		},
	}

	r := chi.NewRouter()
	r.Post("/api/v1/org/{org}/insurance/claims", InsuranceFileClaimHandler(mockStore, mockGHClient))

	claimReq := CreateClaimRequest{
		GithubPRUrl:  "https://github.com/owner/repo/pull/1",
		IncidentDate: time.Now(),
		AmountCents:  1000,
	}
	body, _ := json.Marshal(claimReq)

	req := httptest.NewRequest("POST", "/api/v1/org/"+orgID.String()+"/insurance/claims", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusCreated, rr.Code)

	var claim db.InsuranceClaim
	err := json.NewDecoder(rr.Body).Decode(&claim)
	assert.NoError(t, err)
	assert.Equal(t, orgID, claim.OrgID)
	assert.Equal(t, policyID, claim.PolicyID)
	assert.Equal(t, "PENDING", claim.Status)
}
