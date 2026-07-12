package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

type DependencyPayload struct {
	ProviderRepo         string `json:"provider_repo"`
	ProviderGithubRepoID int64  `json:"provider_github_repo_id"`
	SchemaType           string `json:"schema_type"`
	SpecPath             string `json:"spec_path"`
	Branch               string `json:"branch"`
	RawContent           string `json:"raw_content"`
}

type SyncRequest struct {
	InstallationID       int64               `json:"installation_id"`
	Org                  string              `json:"org"`
	ConsumerRepo         string              `json:"consumer_repo"`
	ConsumerGithubRepoID int64               `json:"consumer_github_repo_id"`
	CommitSHA            string              `json:"commit_sha"`
	Dependencies         []DependencyPayload `json:"dependencies"`
}

// SyncHandler handles the /api/v1/sync endpoint.
func SyncHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		expectedToken := os.Getenv("INTERNAL_SERVICE_TOKEN")
		if expectedToken == "" || authHeader != "Bearer "+expectedToken {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		var req SyncRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
			return
		}

		ctx := r.Context()

		// 1. UpsertOrg
		orgID, err := store.UpsertOrg(ctx, req.InstallationID, req.Org)
		if err != nil {
			http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
			return
		}

		// 2. UpsertRepo for consumer
		parts := strings.Split(req.ConsumerRepo, "/")
		consumerName := req.ConsumerRepo
		if len(parts) == 2 {
			consumerName = parts[1]
		}
		consumerRepoID, err := store.UpsertRepo(ctx, orgID, req.ConsumerGithubRepoID, consumerName, req.ConsumerRepo)
		if err != nil {
			http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
			return
		}

		// 3. For each dependency
		syncedCount := 0
		for _, dep := range req.Dependencies {
			providerParts := strings.Split(dep.ProviderRepo, "/")
			providerName := dep.ProviderRepo
			if len(providerParts) == 2 {
				providerName = providerParts[1]
			}

			providerRepoID, err := store.UpsertRepo(ctx, orgID, dep.ProviderGithubRepoID, providerName, dep.ProviderRepo)
			if err != nil {
				http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
				return
			}

			contractID, err := store.UpsertContract(ctx, providerRepoID, dep.SchemaType, dep.SpecPath, dep.Branch, req.CommitSHA, dep.RawContent)
			if err != nil {
				http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
				return
			}

			// For manual yaml configs, confidence score is implicitly 100 since it is explicitly declared
			if err := store.UpsertDependency(ctx, consumerRepoID, contractID, 100); err != nil {
				http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
				return
			}
			syncedCount++
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int{"synced": syncedCount})
	}
}
