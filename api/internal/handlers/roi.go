package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
	"github.com/go-chi/chi/v5"
)

func ROIHandler(store ports.TelemetryStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := chi.URLParam(r, "org")
		if org == "" {
			http.Error(w, "organization is required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		metrics, err := store.GetROIMetrics(ctx, org)
		if err != nil {
			// We handle missing table or other issues by returning what we have or 500
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(metrics); err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}
