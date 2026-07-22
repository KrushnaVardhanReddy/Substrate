package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
)

func ChangesHandler(store ports.ChangeStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sinceStr := r.URL.Query().Get("since")
		untilStr := r.URL.Query().Get("until")

		if sinceStr == "" {
			http.Error(w, "Missing 'since' parameter", http.StatusBadRequest)
			return
		}

		since, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			http.Error(w, "Invalid 'since' format, expected RFC3339", http.StatusBadRequest)
			return
		}

		until := time.Now().UTC()
		if untilStr != "" {
			parsedUntil, err := time.Parse(time.RFC3339, untilStr)
			if err == nil {
				until = parsedUntil
			} else {
				http.Error(w, "Invalid 'until' format, expected RFC3339", http.StatusBadRequest)
				return
			}
		}

		changes, err := store.GetBreakingChangesBetween(r.Context(), since, until)
		if err != nil {
			http.Error(w, "Failed to fetch breaking changes", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(changes); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
