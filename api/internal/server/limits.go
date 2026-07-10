package server

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/KrushnaVardhanReddy/Substrate/api/internal/db"
)

func TierLimitsMiddleware(store db.Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost || r.Method == http.MethodPut {
				bodyBytes, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, `{"error": "bad request"}`, http.StatusBadRequest)
					return
				}
				r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

				// Parse body loosely to find org, consumer_repo and provider_repo(s).
				// This works for /api/v1/sync and /api/v1/cross-repo-check
				var payload struct {
					Org          string `json:"org"`
					ConsumerRepo string `json:"consumer_repo"`
					ProviderRepo string `json:"provider_repo"` // from cross-repo-check
					Dependencies []struct {
						ProviderRepo string `json:"provider_repo"`
					} `json:"dependencies"` // from sync
				}

				if err := json.Unmarshal(bodyBytes, &payload); err == nil && payload.Org != "" {
					// 1. Max 3 connected repos per org
					repoCount, _ := store.CountReposByOrg(r.Context(), payload.Org)
					repos, _ := store.ListReposByOrg(r.Context(), payload.Org)

					existingSet := make(map[string]bool)
					for _, repo := range repos {
						existingSet[repo.FullName] = true
					}

					newRepos := 0
					if payload.ConsumerRepo != "" && !existingSet[payload.ConsumerRepo] {
						newRepos++
						existingSet[payload.ConsumerRepo] = true
					}
					if payload.ProviderRepo != "" && !existingSet[payload.ProviderRepo] {
						newRepos++
						existingSet[payload.ProviderRepo] = true
					}
					for _, dep := range payload.Dependencies {
						if dep.ProviderRepo != "" && !existingSet[dep.ProviderRepo] {
							newRepos++
							existingSet[dep.ProviderRepo] = true
						}
					}

					if repoCount+newRepos > 3 {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusPaymentRequired)
						json.NewEncoder(w).Encode(map[string]string{
							"error": "Free tier limit reached. Please upgrade to Pro.",
						})
						return
					}

					// 2. Max 5 downstream dependencies per provider
					providersToCheck := make(map[string]bool)
					if payload.ProviderRepo != "" {
						providersToCheck[payload.ProviderRepo] = true
					}
					for _, dep := range payload.Dependencies {
						if dep.ProviderRepo != "" {
							providersToCheck[dep.ProviderRepo] = true
						}
					}

					for provider := range providersToCheck {
						count, _ := store.CountDownstreamDependencies(r.Context(), provider)

						isAlreadyDep := false
						if payload.ConsumerRepo != "" {
							contracts, _ := store.GetContractsByProviderFullName(r.Context(), provider)
							for _, c := range contracts {
								consumers, _ := store.GetConsumersByProviderContract(r.Context(), c.ID)
								for _, cons := range consumers {
									if cons.ConsumerFullName == payload.ConsumerRepo {
										isAlreadyDep = true
										break
									}
								}
								if isAlreadyDep {
									break
								}
							}
						}

						newCount := count
						if !isAlreadyDep && payload.ConsumerRepo != "" {
							newCount++
						}

						if newCount > 5 {
							w.Header().Set("Content-Type", "application/json")
							w.WriteHeader(http.StatusPaymentRequired)
							json.NewEncoder(w).Encode(map[string]string{
								"error": "Free tier limit reached. Please upgrade to Pro.",
							})
							return
						}
					}
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
