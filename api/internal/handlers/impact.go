package handlers

import (
	"encoding/json"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
)

type ImpactResponse struct {
	Provider      string   `json:"provider"`
	RiskScore     int      `json:"risk_score"`
	ImpactedRepos []string `json:"impacted_repos"`
}

// ImpactHandler handles the /api/v1/impact/{org}/{repo} endpoint.
func ImpactHandler(store ports.RepoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := chi.URLParam(r, "org")
		repoName := chi.URLParam(r, "repo")
		if org == "" || repoName == "" {
			http.Error(w, "org and repo missing", http.StatusBadRequest)
			return
		}

		targetProvider := org + "/" + repoName

		edges, err := store.GetDependencyGraph(r.Context(), org)
		if err != nil {
			http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
			return
		}

		// Calculate Nth-degree blast radius.
		// provider -> consumers
		graph := make(map[string][]string)
		for _, edge := range edges {
			graph[edge.ProviderFullName] = append(graph[edge.ProviderFullName], edge.ConsumerFullName)
		}

		impacted := make(map[string]bool)
		queue := []string{targetProvider}

		for len(queue) > 0 {
			curr := queue[0]
			queue = queue[1:]

			for _, consumer := range graph[curr] {
				if !impacted[consumer] {
					impacted[consumer] = true
					queue = append(queue, consumer)
				}
			}
		}

		var impactedList []string
		for repo := range impacted {
			impactedList = append(impactedList, repo)
		}
		sort.Strings(impactedList)
		if impactedList == nil {
			impactedList = []string{}
		}

		resp := ImpactResponse{
			Provider:      targetProvider,
			RiskScore:     len(impactedList),
			ImpactedRepos: impactedList,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}
