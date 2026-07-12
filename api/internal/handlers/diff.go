package handlers

import (
	"encoding/json"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/google/uuid"
	"net/http"
)

type SaveDiffRequest struct {
	DiffReport json.RawMessage `json:"diff_report"`
}

type SaveDiffResponse struct {
	ID string `json:"id"`
}

func SaveDiffHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req SaveDiffRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		if len(req.DiffReport) == 0 {
			http.Error(w, "diff_report is required", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		id, err := store.SaveDiffReport(ctx, req.DiffReport)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(SaveDiffResponse{ID: id.String()})
	}
}

func GetDiffHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := uuid.Parse(idStr)
		if err != nil {
			http.Error(w, "invalid id format", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		diffReport, err := store.GetDiffReport(ctx, id)
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(diffReport)
	}
}
