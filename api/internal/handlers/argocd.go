// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
)

// ArgoWebhookPayload represents the expected payload from ArgoCD/Flux
// Because ArgoCD/Flux webhooks can vary and this is a custom endpoint,
// we will expect a simple JSON payload with org_name and repo_name.
type ArgoWebhookPayload struct {
	OrgName  string `json:"org_name"`
	RepoName string `json:"repo_name"`
}

func ArgoDriftWebhookHandler(store ports.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload ArgoWebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		if payload.OrgName == "" || payload.RepoName == "" {
			http.Error(w, "Missing org_name or repo_name", http.StatusBadRequest)
			return
		}

		anomalies, err := store.GetDriftAnomalies(r.Context(), payload.OrgName, payload.RepoName)
		if err != nil {
			http.Error(w, "Failed to get drift anomalies", http.StatusInternalServerError)
			return
		}

		if len(anomalies) > 0 {
			// Severe schema violations detected, return 406 Not Acceptable to trigger rollback
			http.Error(w, "Severe schema violations detected", http.StatusNotAcceptable)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
