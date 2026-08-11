// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
)

// GraphHandler handles the /api/v1/graph/{org} endpoint.
func GraphHandler(store ports.RepoStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := r.PathValue("org")
		if org == "" {
			http.Error(w, "org missing", http.StatusBadRequest)
			return
		}

		edges, err := store.GetDependencyGraph(r.Context(), org)
		if err != nil {
			http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
			return
		}

		if edges == nil {
			edges = []db.DependencyEdge{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(edges)
	}
}
