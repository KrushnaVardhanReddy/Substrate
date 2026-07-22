package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
	"github.com/go-chi/chi/v5"
)

const badgeTemplate = `<svg xmlns="http://www.w3.org/2000/svg" width="300" height="20">
  <linearGradient id="b" x2="0" y2="100%%">
    <stop offset="0" stop-color="#bbb" stop-opacity=".1"/>
    <stop offset="1" stop-opacity=".1"/>
  </linearGradient>
  <mask id="a">
    <rect width="300" height="20" rx="3" fill="#fff"/>
  </mask>
  <g mask="url(#a)">
    <path fill="#555" d="M0 0h90v20H0z"/>
    <path fill="%s" d="M90 0h210v20H90z"/>
    <path fill="url(#b)" d="M0 0h300v20H0z"/>
  </g>
  <g fill="#fff" text-anchor="middle" font-family="DejaVu Sans,Verdana,Geneva,sans-serif" font-size="11">
    <text x="45" y="15" fill="#010101" fill-opacity=".3">CONTRACT</text>
    <text x="45" y="14">CONTRACT</text>
    <text x="195" y="15" fill="#010101" fill-opacity=".3">%s</text>
    <text x="195" y="14">%s</text>
  </g>
</svg>`

func calculateScore(breaks int) (string, string) {
	if breaks == 0 {
		return "A+ | 0 breaks in 90 days", "#4c1" // brightgreen
	} else if breaks <= 2 {
		return fmt.Sprintf("B | %d breaks in 90 days", breaks), "#dfb317" // yellow
	} else if breaks <= 5 {
		return fmt.Sprintf("C | %d breaks in 90 days", breaks), "#e05d44" // red
	} else {
		return fmt.Sprintf("F | %d breaks in 90 days", breaks), "#e05d44" // red
	}
}

func BadgesHandler(store ports.BadgesStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := chi.URLParam(r, "org")
		repo := chi.URLParam(r, "repo")

		if org == "" || repo == "" {
			http.Error(w, "missing org or repo", http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		repos, err := store.ListReposByOrg(ctx, org)
		if err != nil {
			http.Error(w, "failed to fetch repos", http.StatusInternalServerError)
			return
		}

		var repoID *db.Repository
		for _, r := range repos {
			if r.Name == repo {
				repoID = &r
				break
			}
		}

		if repoID == nil {
			http.Error(w, "repo not found", http.StatusNotFound)
			return
		}

		since := time.Now().Add(-90 * 24 * time.Hour)
		breaks, err := store.CountRecentBreakingChanges(ctx, repoID.ID, since)
		if err != nil {
			http.Error(w, "failed to get breaking changes", http.StatusInternalServerError)
			return
		}

		scoreText, color := calculateScore(breaks)

		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "public, max-age=86400") // 24-hour TTL

		svg := fmt.Sprintf(badgeTemplate, color, scoreText, scoreText)
		_, _ = w.Write([]byte(svg))
	}
}
