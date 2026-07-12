package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

type CanDeployResponse struct {
	Safe      bool     `json:"safe"`
	Message   string   `json:"message"`
	BlockedBy []string `json:"blocked_by,omitempty"`
}

func sendJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(CanDeployResponse{
		Safe:    false,
		Message: message,
	})
}

func CanDeployHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := r.URL.Query().Get("repo")
		if repo == "" {
			sendJSONError(w, "missing repo parameter", http.StatusBadRequest)
			return
		}
		commit := r.URL.Query().Get("commit")
		if commit == "" {
			sendJSONError(w, "missing commit parameter", http.StatusBadRequest)
			return
		}

		parts := strings.Split(repo, "/")
		if len(parts) != 2 {
			sendJSONError(w, "invalid repo format, expected owner/repo", http.StatusBadRequest)
			return
		}
		orgName := parts[0]
		repoName := parts[1]

		ctx := r.Context()

		breakingHistory, err := store.GetBreakingChangeHistory(ctx, orgName, repoName, 1)
		if err != nil {
			sendJSONError(w, "failed to get breaking change history", http.StatusInternalServerError)
			return
		}

		var hasBreakingChange bool
		for _, bc := range breakingHistory {
			if bc.GitSHA == commit {
				hasBreakingChange = true
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if !hasBreakingChange {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(CanDeployResponse{
				Safe:    true,
				Message: "Safe to deploy: no breaking change found",
			})
			return
		}

		edges, err := store.GetDependencyGraph(ctx, orgName)
		if err != nil {
			sendJSONError(w, "failed to get dependency graph", http.StatusInternalServerError)
			return
		}

		var blockedBy []string
		seen := make(map[string]bool)
		for _, edge := range edges {
			if edge.ProviderFullName == repo {
				if !seen[edge.ConsumerFullName] {
					seen[edge.ConsumerFullName] = true
					blockedBy = append(blockedBy, edge.ConsumerFullName)
				}
			}
		}

		if len(blockedBy) > 0 {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(CanDeployResponse{
				Safe:      false,
				Message:   "Deployment blocked: required consumers have not updated to handle breaking change",
				BlockedBy: blockedBy,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CanDeployResponse{
			Safe:    true,
			Message: "Safe to deploy",
		})
	}
}
