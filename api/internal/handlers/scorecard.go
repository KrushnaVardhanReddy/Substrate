// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/scoring"
)

// GetScorecardHandler computes and returns the scorecard for a specific org/repo
func GetScorecardHandler(store ports.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := r.PathValue("org")
		repo := r.PathValue("repo")

		if org == "" || repo == "" {
			http.Error(w, `{"error": "org or repo missing"}`, http.StatusBadRequest)
			return
		}

		scorecard, err := scoring.ComputeScore(r.Context(), store, org, repo)
		if err != nil {
			http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(scorecard)
	}
}
