package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/services"
)

type CanRollbackResponse struct {
	CanRollback bool     `json:"can_rollback"`
	Message     string   `json:"message"`
	BlockedBy   []string `json:"blocked_by,omitempty"`
}

func sendRollbackJSONError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(CanRollbackResponse{
		CanRollback: false,
		Message:     message,
	})
}

func CanRollbackHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := r.URL.Query().Get("org")
		repo := r.URL.Query().Get("repo")
		targetSha := r.URL.Query().Get("target_sha")

		if org == "" || repo == "" || targetSha == "" {
			sendRollbackJSONError(w, "missing required parameters", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		providerFullName := org + "/" + repo

		contracts, err := store.GetContractsByProviderFullName(ctx, providerFullName)
		if err != nil {
			sendRollbackJSONError(w, "failed to get contracts", http.StatusInternalServerError)
			return
		}

		var targetContractContent string
		var targetSchemaType string
		foundTarget := false

		for _, c := range contracts {
			if c.LatestCommitSHA == targetSha {
				targetContractContent = c.RawContent
				targetSchemaType = c.SchemaType
				foundTarget = true
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")

		if !foundTarget {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(CanRollbackResponse{
				CanRollback: true,
				Message:     "Target commit SHA not found in contracts, assuming safe to rollback",
			})
			return
		}

		req := services.CrossRepoCheckRequest{
			InstallationID:    0,
			Org:               org,
			ProviderRepo:      providerFullName,
			HeadSchemaContent: targetContractContent,
			SchemaType:        targetSchemaType,
		}

		checkResponse, err := services.PerformCrossRepoCheck(ctx, store, req)
		if err != nil {
			sendRollbackJSONError(w, "failed to perform cross repo check", http.StatusInternalServerError)
			return
		}

		var blockedBy []string
		for _, res := range checkResponse.Results {
			if res.Status == "breaking" || res.Status == "error" {
				blockedBy = append(blockedBy, res.ConsumerRepo)
			}
		}

		if len(blockedBy) > 0 {
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(CanRollbackResponse{
				CanRollback: false,
				Message:     "Rollback blocked: downstream consumers rely on the current schema and would break",
				BlockedBy:   blockedBy,
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CanRollbackResponse{
			CanRollback: true,
			Message:     "Safe to rollback",
		})
	}
}
