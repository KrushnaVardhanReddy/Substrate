package registry

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/go-chi/chi/v5"
)

type PublishRequest struct {
	SchemaType    string `json:"schema_type"`
	SchemaContent string `json:"schema_content"`
}

func HandlePublishSchema(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := chi.URLParam(r, "namespace")
		name := chi.URLParam(r, "name")
		version := chi.URLParam(r, "version")

		if namespace == "" || name == "" || version == "" {
			http.Error(w, "missing parameters", http.StatusBadRequest)
			return
		}

		var req PublishRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.SchemaType == "" || req.SchemaContent == "" {
			http.Error(w, "schema_type and schema_content are required", http.StatusBadRequest)
			return
		}

		if err := store.PublishPublicSchema(r.Context(), namespace, name, version, req.SchemaType, req.SchemaContent); err != nil {
			http.Error(w, "failed to publish schema", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func HandleFetchSchema(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := chi.URLParam(r, "namespace")
		name := chi.URLParam(r, "name")
		version := chi.URLParam(r, "version")

		if namespace == "" || name == "" || version == "" {
			http.Error(w, "missing parameters", http.StatusBadRequest)
			return
		}

		schema, err := store.GetPublicSchema(r.Context(), namespace, name, version)
		if err != nil {
			// In a real app we'd check if it's a "not found" error vs internal error.
			http.Error(w, "schema not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(schema)
	}
}
