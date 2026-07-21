package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

func SpecHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := r.PathValue("org")
		repo := r.PathValue("repo")

		if org == "" || repo == "" {
			http.Error(w, "org and repo are required", http.StatusBadRequest)
			return
		}

		fullName := org + "/" + repo

		contracts, err := store.GetContractsByProviderFullName(r.Context(), fullName)
		if err != nil {
			http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
			return
		}

		if len(contracts) == 0 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "no spec found for repo"})
			return
		}

		c := contracts[0]

		response := map[string]string{
			"spec": c.RawContent,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
