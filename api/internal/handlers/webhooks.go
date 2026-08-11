// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
	"github.com/go-chi/chi/v5"
)

type RegisterWebhookRequest struct {
	URL    string `json:"url"`
	Secret string `json:"secret"`
}

func RegisterWebhookHandler(store ports.WebhookStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := chi.URLParam(r, "org")
		if org == "" {
			http.Error(w, "organization is required", http.StatusBadRequest)
			return
		}

		var req RegisterWebhookRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.URL == "" {
			http.Error(w, "url is required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		config := db.WebhookConfig{
			Org:    org,
			URL:    req.URL,
			Secret: req.Secret,
		}

		err := store.RegisterWebhook(ctx, config)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}
