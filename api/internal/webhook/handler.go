package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/Substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/Substrate/api/internal/discovery"
	"github.com/KrushnaVardhanReddy/Substrate/api/internal/integrations/postman"
)

type PushPayload struct {
	InstallationID int64  `json:"installation_id"`
	Org            string `json:"org"`
	Repo           string `json:"repo"`
	GithubRepoID   int64  `json:"github_repo_id"`
	CommitSHA      string `json:"commit_sha"`
	Files          []File `json:"files"`
}

type File struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// PushHandler handles the GitHub push webhook payload forwarded from the GitHub App
func PushHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PushPayload
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid json"}`, http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		orgID, err := store.UpsertOrg(ctx, req.InstallationID, req.Org)
		if err != nil {
			http.Error(w, `{"error": "internal error on UpsertOrg"}`, http.StatusInternalServerError)
			return
		}

		consumerName := req.Repo
		parts := strings.Split(req.Repo, "/")
		if len(parts) == 2 {
			consumerName = parts[1]
		}

		consumerRepoID, err := store.UpsertRepo(ctx, orgID, req.GithubRepoID, consumerName, req.Repo)
		if err != nil {
			http.Error(w, `{"error": "internal error on UpsertRepo"}`, http.StatusInternalServerError)
			return
		}

		discoveredCount := 0
		for _, file := range req.Files {
			var deps []discovery.DiscoveredDependency

			if strings.HasSuffix(file.Path, ".env.example") || strings.HasSuffix(file.Path, ".env.template") || strings.HasSuffix(file.Path, ".env.sample") {
				deps = append(deps, discovery.ScanEnvFile(file.Content)...)
			} else if strings.Contains(file.Path, "docker-compose") {
				deps = append(deps, discovery.ScanDockerCompose(file.Content)...)
			} else if strings.HasSuffix(file.Path, ".yaml") || strings.HasSuffix(file.Path, ".yml") {
				deps = append(deps, discovery.ScanKubernetesManifest(file.Content)...)
			}

			for _, dep := range deps {
				providerRepoName := extractProviderFromURL(dep.VarValue)
				if providerRepoName == "" {
					continue
				}

				providerRepoID, err := store.UpsertRepo(ctx, orgID, 0, providerRepoName, req.Org+"/"+providerRepoName)
				if err != nil {
					continue
				}

				contractID, err := store.UpsertContract(ctx, providerRepoID, "unknown", "discovered", req.CommitSHA, req.CommitSHA, dep.VarValue)
				if err != nil {
					continue
				}

				err = store.UpsertDependency(ctx, consumerRepoID, contractID, dep.ConfidenceScore)
				if err == nil {
					discoveredCount++
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int{"discovered": discoveredCount})
	}
}

func extractProviderFromURL(urlStr string) string {
	urlStr = strings.TrimPrefix(urlStr, "http://")
	urlStr = strings.TrimPrefix(urlStr, "https://")

	parts := strings.Split(urlStr, ":")
	host := parts[0]

	hostParts := strings.Split(host, ".")
	if len(hostParts) > 0 {
		return hostParts[0]
	}
	return ""
}

type WebhookPayload struct {
	Event   string `json:"event"`
	Status  string `json:"status"` // "SAFE", "APPROVED", etc.
	Schema  string `json:"schema"` // The OpenAPI schema JSON/YAML string
	Type    string `json:"type"`   // "openapi", etc.
	Version string `json:"version"`
}

type SyncClient interface {
	SyncSchema(ctx context.Context, schema string) error
}

func Handler() http.HandlerFunc {
	return HandlerWithClient(postman.NewClient())
}

func HandlerWithClient(client SyncClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload WebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		if payload.Type == "openapi" && (payload.Status == "SAFE" || payload.Status == "APPROVED") {
			ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()
			err := client.SyncSchema(ctx, payload.Schema)
			if err != nil {
				http.Error(w, "failed to sync to postman", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
