package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/KrushnaVardhanReddy/Substrate/api/internal/db"
)

func SchemaHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		owner := r.PathValue("owner")
		repo := r.PathValue("repo")

		if owner == "" || repo == "" {
			http.Error(w, "owner and repo are required", http.StatusBadRequest)
			return
		}

		fullName := owner + "/" + repo

		contracts, err := store.GetContractsByProviderFullName(r.Context(), fullName)
		if err != nil {
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
			return
		}

		if len(contracts) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "no schema found for repo"})
			return
		}

		c := contracts[0]

		type schemaResponse struct {
			Schema     string `json:"schema"`
			SchemaType string `json:"schema_type"`
			SpecPath   string `json:"spec_path"`
			CommitSHA  string `json:"commit_sha"`
			SyncedAt   string `json:"synced_at"`
		}

		response := schemaResponse{
			Schema:     c.RawContent,
			SchemaType: c.SchemaType,
			SpecPath:   c.SpecPath,
			CommitSHA:  c.LatestCommitSHA,
			SyncedAt:   c.SyncedAt.Format(time.RFC3339),
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
