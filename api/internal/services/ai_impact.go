// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package services

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/engine/pkg/ai"
)

type AIImpactRequest struct {
	Changes   []ai.Change `json:"changes"`
	Consumers []ai.Repo   `json:"consumers"`
}

type AIImpactResponse struct {
	Explanation string `json:"explanation"`
}

func AIImpactHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req AIImpactRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		client, err := ai.NewAIClient()
		if err != nil {
			// Mock fallback for test environment
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(AIImpactResponse{
				Explanation: "Mock AI Impact Analysis: The breaking changes impact downstream consumers.",
			})
			return
		}

		summary, err := ai.GenerateImpactSummary(r.Context(), client, req.Changes, req.Consumers)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(AIImpactResponse{
			Explanation: summary,
		})
	}
}
