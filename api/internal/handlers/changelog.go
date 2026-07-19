package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

type ChangelogEntry struct {
	Date            string `json:"date"`
	EndpointsAdded  int    `json:"endpoints_added"`
	EndpointsRemoved int   `json:"endpoints_removed"`
}

func ChangelogHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := chi.URLParam(r, "org")
		repo := chi.URLParam(r, "repo")

		if org == "" || repo == "" {
			http.Error(w, "missing org or repo", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		reports, err := store.GetDiffReportsByRepo(ctx, org, repo, 20)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		var changelog []ChangelogEntry
		for _, record := range reports {
			var report struct {
				Summary    struct {
					BreakingCount int `json:"breaking_count"`
					SafeCount     int `json:"safe_count"` // Adding is usually safe
				} `json:"summary"`
				BreakingChanges []struct {
					RuleID string `json:"rule_id"`
				} `json:"breaking_changes"`
				SafeChanges []struct {
					RuleID string `json:"rule_id"`
				} `json:"safe_changes"`
			}

			if err := json.Unmarshal(record.ReportData, &report); err != nil {
				continue
			}

			var added, removed int
			for _, bc := range report.BreakingChanges {
				if bc.RuleID == "ENDPOINT_REMOVED" || bc.RuleID == "METHOD_REMOVED" {
					removed++
				}
			}
			for _, sc := range report.SafeChanges {
				if sc.RuleID == "ENDPOINT_ADDED" || sc.RuleID == "METHOD_ADDED" {
					added++
				}
			}

			// Even if 0 added and 0 removed, we might still include it if there's other stuff,
			// but for a summary, let's include it anyway, or maybe just if there's any change?
			// The requirements say: "chronological summary JSON array containing dates, endpoints added, and endpoints removed."
			changelog = append(changelog, ChangelogEntry{
				Date:            record.CreatedAt.Format("2006-01-02T15:04:05Z"),
				EndpointsAdded:  added,
				EndpointsRemoved: removed,
			})
		}

		if changelog == nil {
			changelog = []ChangelogEntry{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(changelog)
	}
}
