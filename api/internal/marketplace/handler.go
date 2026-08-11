// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package marketplace

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

func PublishHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Name        string          `json:"name"`
			Description string          `json:"description"`
			Schema      json.RawMessage `json:"schema_content"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.Name == "" || len(req.Schema) == 0 {
			http.Error(w, "name and schema_content are required", http.StatusBadRequest)
			return
		}

		id, err := store.PublishPlugin(r.Context(), req.Name, req.Description, req.Schema)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"id": id.String()})
	}
}

func ListPluginsHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		plugins, err := store.ListPlugins(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if plugins == nil {
			plugins = []db.MarketplacePlugin{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(plugins)
	}
}
