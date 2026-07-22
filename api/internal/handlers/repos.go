package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
)

// ReposHandler handles the /api/v1/repos/{org} endpoint.
func ReposHandler(store ports.RepoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := r.PathValue("org")
		if org == "" {
			http.Error(w, "org missing", http.StatusBadRequest)
			return
		}

		repos, err := store.ListReposByOrg(r.Context(), org)
		if err != nil {
			http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
			return
		}

		if repos == nil {
			repos = []db.Repository{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(repos)
	}
}
