package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/go-chi/chi/v5"
)

// ListGuidesHandler returns the list of guide documents for a repo.
func ListGuidesHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := chi.URLParam(r, "org")
		repo := chi.URLParam(r, "repo")

		guides, err := store.ListRepoGuides(r.Context(), org, repo)
		if err != nil {
			http.Error(w, `{"error": "failed to list guides"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(guides)
	}
}

// GetGuideHandler returns a single rendered guide's content.
func GetGuideHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := chi.URLParam(r, "org")
		repo := chi.URLParam(r, "repo")
		slug := chi.URLParam(r, "slug")

        // Since slug might contain slashes if it's deeply nested in docs/, we should get the rest of the url
        slugPath := chi.URLParam(r, "*")
        if slugPath != "" {
            slug = slug + "/" + slugPath
        }

		guide, err := store.GetRepoGuide(r.Context(), org, repo, slug)
		if err != nil {
			http.Error(w, `{"error": "guide not found"}`, http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(guide)
	}
}
