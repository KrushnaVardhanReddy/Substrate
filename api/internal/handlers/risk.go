package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/KrushnaVardhanReddy/substrate/engine/risk"
)

type RiskScoreResponse struct {
	Score string `json:"score"`
}

type RiskScoreRequest struct {
	BreakingChangesCount int  `json:"breaking_changes_count"`
	BlastRadiusNodeCount int  `json:"blast_radius_node_count"`
	HasDBMigrations      bool `json:"has_db_migrations"`
	E2ETestsPass         bool `json:"e2e_tests_pass"`
}

func RiskScoreHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := r.PathValue("org")
		repo := r.PathValue("repo")
		pr := r.PathValue("pr")

		if org == "" || repo == "" || pr == "" {
			http.Error(w, "missing required path parameters", http.StatusBadRequest)
			return
		}

		breakingChangesCount := 0
		blastRadiusNodeCount := 0
		hasDBMigrations := false
		e2eTestsPass := true

		if val := r.URL.Query().Get("breaking_changes_count"); val != "" {
			if parsed, err := strconv.Atoi(val); err == nil {
				breakingChangesCount = parsed
			}
		}

		if val := r.URL.Query().Get("blast_radius_node_count"); val != "" {
			if parsed, err := strconv.Atoi(val); err == nil {
				blastRadiusNodeCount = parsed
			}
		}

		if val := r.URL.Query().Get("has_db_migrations"); val != "" {
			if parsed, err := strconv.ParseBool(val); err == nil {
				hasDBMigrations = parsed
			}
		}

		if val := r.URL.Query().Get("e2e_tests_pass"); val != "" {
			if parsed, err := strconv.ParseBool(val); err == nil {
				e2eTestsPass = parsed
			}
		}

		scoreInput := risk.ScoreInput{
			BreakingChangesCount: breakingChangesCount,
			BlastRadiusNodeCount: blastRadiusNodeCount,
			HasDBMigrations:      hasDBMigrations,
			E2ETestsPass:         e2eTestsPass,
		}

		score := risk.CalculateRiskScore(scoreInput)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RiskScoreResponse{Score: score})
	}
}
