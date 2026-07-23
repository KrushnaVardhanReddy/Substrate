package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type CreateClaimRequest struct {
	GithubPRUrl  string    `json:"github_pr_url"`
	IncidentDate time.Time `json:"incident_date"`
	AmountCents  int64     `json:"amount_cents"`
}

func InsuranceGetPolicyHandler(store ports.InsuranceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgName := chi.URLParam(r, "org")
		ctx := r.Context()
		orgID, err := store.GetOrgIDByName(ctx, orgName)
		if err != nil {
			http.Error(w, "invalid org", http.StatusBadRequest)
			return
		}

		policy, err := store.GetInsurancePolicy(ctx, orgID)
		if err != nil {
			http.Error(w, "insurance policy not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(policy)
	}
}

func InsuranceGetClaimsHandler(store ports.InsuranceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgName := chi.URLParam(r, "org")
		ctx := r.Context()
		orgID, err := store.GetOrgIDByName(ctx, orgName)
		if err != nil {
			http.Error(w, "invalid org", http.StatusBadRequest)
			return
		}

		claims, err := store.GetInsuranceClaims(ctx, orgID)
		if err != nil {
			http.Error(w, "failed to query claims", http.StatusInternalServerError)
			return
		}

		if claims == nil {
			claims = make([]db.InsuranceClaim, 0)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(claims)
	}
}

func InsuranceFileClaimHandler(store ports.InsuranceStore, githubClient github.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgName := chi.URLParam(r, "org")
		ctx := r.Context()
		orgID, err := store.GetOrgIDByName(ctx, orgName)
		if err != nil {
			http.Error(w, "invalid org", http.StatusBadRequest)
			return
		}

		var req CreateClaimRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Get the policy
		policy, err := store.GetInsurancePolicy(ctx, orgID)
		if err != nil {
			http.Error(w, "no active insurance policy found for this organization", http.StatusNotFound)
			return
		}

		status := "PENDING"

		parts := strings.Split(strings.TrimPrefix(req.GithubPRUrl, "https://github.com/"), "/")
		if len(parts) >= 4 && parts[2] == "pull" {
			owner := parts[0]
			repo := parts[1]

			// Extract PR number
			prNumStr := parts[3]
			prNum, err := strconv.Atoi(prNumStr)
			if err == nil {
				headSHA, err := githubClient.GetPullRequestHeadSHA(ctx, owner, repo, prNum)
				if err == nil {
					checkRuns, err := githubClient.ListCheckRunsForRef(ctx, owner, repo, headSHA)
					if err == nil {
						for _, run := range checkRuns {
							if strings.HasPrefix(run.Name, "substrate") {
								if run.Conclusion == "neutral" || run.Conclusion == "success" { // If overridden or success but still broke
									status = "REJECTED"
								}
							}
						}
					}
				}
			}
		}

		claimID := uuid.New()
		now := time.Now()

		claim := db.InsuranceClaim{
			ID:           claimID,
			OrgID:        orgID,
			PolicyID:     policy.ID,
			GithubPRUrl:  req.GithubPRUrl,
			IncidentDate: req.IncidentDate,
			Status:       status,
			AmountCents:  req.AmountCents,
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		_, err = store.CreateInsuranceClaim(ctx, claim)
		if err != nil {
			http.Error(w, "failed to create claim", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(claim)
	}
}
