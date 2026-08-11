// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
	"github.com/google/uuid"
)

type RecordBreakingChangeRequest struct {
	RepoID          uuid.UUID       `json:"repo_id"`
	OrgName         string          `json:"org_name"`
	RepoName        string          `json:"repo_name"`
	GitSHA          string          `json:"git_sha"`
	BreakingChanges json.RawMessage `json:"breaking_changes"`
}

func HistoryHandler(store ports.ChangeStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read body", http.StatusBadRequest)
				return
			}
			defer r.Body.Close()

			var req RecordBreakingChangeRequest
			if err := json.Unmarshal(body, &req); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			if req.RepoID == uuid.Nil || req.OrgName == "" || req.RepoName == "" || req.GitSHA == "" || len(req.BreakingChanges) == 0 {
				http.Error(w, "Missing required fields", http.StatusBadRequest)
				return
			}

			err = store.RecordBreakingChange(r.Context(), req.RepoID, req.OrgName, req.RepoName, req.GitSHA, req.BreakingChanges)
			if err != nil {
				http.Error(w, "Failed to record breaking change", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func HistoryGetHandler(store ports.ChangeStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		org := r.PathValue("org")
		repo := r.PathValue("repo")

		if org == "" || repo == "" {
			http.Error(w, "Missing org or repo", http.StatusBadRequest)
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := 10
		if limitStr != "" {
			parsedLimit, err := strconv.Atoi(limitStr)
			if err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		history, err := store.GetBreakingChangeHistory(r.Context(), org, repo, limit)
		if err != nil {
			http.Error(w, "Failed to fetch history", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(history); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
	}
}
