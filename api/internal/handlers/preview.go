package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func GetPreviewHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := chi.URLParam(r, "token")
		token, err := uuid.Parse(tokenStr)
		if err != nil {
			http.Error(w, "invalid token format", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		diffReport, expiresAt, err := store.GetPreviewSession(ctx, token)
		if err != nil {
			http.Error(w, "preview session not found", http.StatusNotFound)
			return
		}

		if time.Now().After(expiresAt) {
			http.Error(w, "preview session expired", http.StatusGone)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(diffReport)
	}
}

type ExpirePreviewRequest struct {
	Org      string `json:"org"`
	Repo     string `json:"repo"`
	PRNumber int    `json:"pr_number"`
}

func ExpirePreviewHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req ExpirePreviewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if req.Org == "" || req.Repo == "" || req.PRNumber == 0 {
			http.Error(w, "missing required fields", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		err := store.ExpirePreviewSessionsForPR(ctx, req.PRNumber, req.Org, req.Repo)
		if err != nil {
			http.Error(w, "failed to expire preview sessions", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success"})
	}
}
