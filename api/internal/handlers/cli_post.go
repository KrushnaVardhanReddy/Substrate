// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/engine/linter"
	"github.com/KrushnaVardhanReddy/substrate/engine/postmortem"
)

// PostmortemHandler handles /api/v1/postmortem requests.
func PostmortemHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			IncidentDate string `json:"incident_date"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if req.IncidentDate == "" {
			http.Error(w, "incident_date is required", http.StatusBadRequest)
			return
		}

		incidentDate, err := time.Parse(time.RFC3339, req.IncidentDate)
		if err != nil {
			http.Error(w, "invalid incident date format", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/markdown")

		scheme := "http"
		if r.TLS != nil {
			scheme = "https"
		}
		apiURL := scheme + "://" + r.Host

		token := r.Header.Get("Authorization")
		if len(token) > 7 && token[:7] == "Bearer " {
			token = token[7:]
		}

		err = postmortem.GeneratePostMortem(r.Context(), apiURL, token, incidentDate, w)
		if err != nil {
			return
		}
	}
}

// SchemaSmellHandler handles /api/v1/schema/smell requests.
func SchemaSmellHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		schemaContent, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if len(schemaContent) == 0 {
			http.Error(w, "schema body is required", http.StatusBadRequest)
			return
		}

		score, issues, err := linter.Analyze(r.Context(), schemaContent)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		out := map[string]interface{}{
			"score":  score,
			"issues": issues,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
	}
}
