// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PublicProfileHandler struct {
	pool *pgxpool.Pool
}

func NewPublicProfileHandler(pool *pgxpool.Pool) *PublicProfileHandler {
	return &PublicProfileHandler{pool: pool}
}

func (h *PublicProfileHandler) GetPublicProfile(w http.ResponseWriter, r *http.Request) {
	orgName := chi.URLParam(r, "org")

	var isOSS bool
	err := h.pool.QueryRow(r.Context(), `
		SELECT is_oss FROM organizations WHERE github_org_name = $1
	`, orgName).Scan(&isOSS)

	if err != nil || !isOSS {
		http.Error(w, "Organization not found or not OSS", http.StatusNotFound)
		return
	}

	// Mocking the data for the 90-day reliability metrics
	now := time.Now()
	type MetricPoint struct {
		Date  string `json:"date"`
		Value int    `json:"value"`
	}

	var breakingChanges []MetricPoint
	var blastRadius []MetricPoint
	var uptime []MetricPoint

	for i := 89; i >= 0; i-- {
		dateStr := now.AddDate(0, 0, -i).Format("2006-01-02")
		breakingChanges = append(breakingChanges, MetricPoint{Date: dateStr, Value: 0})
		blastRadius = append(blastRadius, MetricPoint{Date: dateStr, Value: 10}) // Mock static blast radius
		uptime = append(uptime, MetricPoint{Date: dateStr, Value: 100})          // Mock 100% uptime
	}

	response := map[string]interface{}{
		"org_name": orgName,
		"metrics": map[string]interface{}{
			"breaking_changes": breakingChanges,
			"blast_radius":     blastRadius,
			"uptime":           uptime,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
