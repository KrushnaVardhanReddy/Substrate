// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
	"github.com/go-chi/chi/v5"
)

func GetDriftReportHandler(store ports.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := chi.URLParam(r, "org")
		repo := chi.URLParam(r, "repo")

		if org == "" || repo == "" {
			http.Error(w, "missing org or repo", http.StatusBadRequest)
			return
		}

		// Typically we would fetch from store here. For sidecar testing we can return the mock format
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "drift_detected",
			"report": "Mock sidecar drift report for " + org + "/" + repo,
		})
	}
}
