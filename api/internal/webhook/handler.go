package webhook

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/KrushnaVardhanReddy/Substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/Substrate/api/internal/discovery"
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
				// We don't have a provider registry yet to look up the URL.
				// We need to create a dummy provider contract for now, or see if it exists.
				// Wait, if it doesn't exist, we can't create a dependency!
				// What does phase-5 spec say?
				// "resolve against URL->Repo Registry"
				// Since we don't have the full registry in this task (Tier 1 Env Scanner),
				// we will just look for existing contracts where the provider repo name might match the URL roughly,
				// or just insert the dependency if we can resolve it.
				// Let's resolve the URL to a repo string by stripping http and matching repo name.

				providerRepoName := extractProviderFromURL(dep.VarValue)
				if providerRepoName == "" {
					continue
				}

				// Provide dummy github repo ID since we don't know it
				providerRepoID, err := store.UpsertRepo(ctx, orgID, 0, providerRepoName, req.Org+"/"+providerRepoName)
				if err != nil {
					continue
				}

				// Create a placeholder contract
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
	// A naive extractor for now: https://users.myorg.com -> users
	// http://users-service:8080 -> users-service
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
