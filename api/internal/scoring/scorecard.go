package scoring

import (
	"context"
	"encoding/json"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
	"github.com/jackc/pgx/v5"
)

type ScorecardBreakdown struct {
	Docs       int `json:"docs"`
	Compliance int `json:"compliance"`
	Volatility int `json:"volatility"`
	Coverage   int `json:"coverage"`
}

type Scorecard struct {
	Grade     string             `json:"grade"`
	Total     int                `json:"total"`
	Breakdown ScorecardBreakdown `json:"breakdown"`
}

// ComputeScore evaluates documentation quality, compliance risk, volatility, and QA test coverage
// to calculate an API Health Grade (A-F).
func ComputeScore(ctx context.Context, store ports.Store, org, repo string) (*Scorecard, error) {
	// 1. Check cache (1-hour TTL)
	cached, err := store.GetScorecardCache(ctx, sqlcgen.GetScorecardCacheParams{
		Org:  org,
		Repo: repo,
	})
	if err == nil {
		if time.Since(cached.ComputedAt) < time.Hour {
			var breakdown ScorecardBreakdown
			if err := json.Unmarshal(cached.Breakdown, &breakdown); err == nil {
				return &Scorecard{
					Grade:     cached.Grade,
					Total:     int(cached.Total),
					Breakdown: breakdown,
				}, nil
			}
		}
	} else if err != pgx.ErrNoRows {
		// Log the error but proceed to compute if not just a cache miss
	}

	breakdown := ScorecardBreakdown{
		Docs:       0,
		Compliance: 25,
		Volatility: 25,
		Coverage:   0,
	}

	// 2. Documentation completeness (existence of repo guides)
	hasGuides, err := store.HasRepoGuides(ctx, sqlcgen.HasRepoGuidesParams{
		Org:  org,
		Repo: repo,
	})
	if err == nil && hasGuides {
		breakdown.Docs = 25
	}

	// 3. Compliance risk (from repo_compliance)
	compliance, err := store.GetRepoCompliance(ctx, sqlcgen.GetRepoComplianceParams{
		Org:  org,
		Repo: repo,
	})
	if err == nil {
		if compliance.HasPii {
			breakdown.Compliance -= 5
		}
		if !compliance.Soc2Compliant {
			breakdown.Compliance -= 5
		}
	} else if err == pgx.ErrNoRows {
		// Default to not compliant if missing
		breakdown.Compliance -= 5 // Missing SOC2
	}

	// 4. Volatility (from diff_reports in last 30 days)
	limit := 100 // reasonable upper bound
	reports, err := store.GetDiffReportsByRepo(ctx, org, repo, limit)
	if err == nil {
		breakingCount := 0
		for _, rep := range reports {
			if time.Since(rep.CreatedAt) > 30*24*time.Hour {
				break // Reports are ordered by descending created_at
			}
			// Unmarshal and count breaking changes
			var diffRep struct {
				Summary struct {
					BreakingCount int `json:"breaking_count"`
				} `json:"summary"`
			}
			if err := json.Unmarshal(rep.ReportData, &diffRep); err == nil {
				breakingCount += diffRep.Summary.BreakingCount
			}
		}

		volatilityDeduction := breakingCount * 3
		breakdown.Volatility -= volatilityDeduction
		if breakdown.Volatility < 0 {
			breakdown.Volatility = 0
		}
	}

	// 5. QA shadow coverage (from qa_replay_jobs)
	coverageScore, err := store.GetLatestQACoverageScore(ctx, sqlcgen.GetLatestQACoverageScoreParams{
		GithubOrgName: org,
		Name:          repo,
	})
	if err == nil {
		val, err := coverageScore.Float64Value()
		if err == nil && val.Valid {
			scaled := (val.Float64 / 100) * 25
			breakdown.Coverage = int(scaled)
		} else {
			// Fallback if type casting fails or is invalid.
			// Attempt retrieving integer representation.
			intVal, err2 := coverageScore.Int64Value()
			if err2 == nil && intVal.Valid {
				scaled := (float64(intVal.Int64) / 100) * 25
				breakdown.Coverage = int(scaled)
			}
		}
	}

	// Compute total & grade
	total := breakdown.Docs + breakdown.Compliance + breakdown.Volatility + breakdown.Coverage
	grade := calculateGrade(total)

	breakdownJSON, _ := json.Marshal(breakdown)

	// Save to cache
	_ = store.UpsertScorecardCache(ctx, sqlcgen.UpsertScorecardCacheParams{
		Org:       org,
		Repo:      repo,
		Grade:     grade,
		Total:     int32(total),
		Breakdown: breakdownJSON,
	})

	return &Scorecard{
		Grade:     grade,
		Total:     total,
		Breakdown: breakdown,
	}, nil
}

func calculateGrade(total int) string {
	if total >= 90 {
		return "A"
	}
	if total >= 75 {
		return "B"
	}
	if total >= 60 {
		return "C"
	}
	if total >= 40 {
		return "D"
	}
	return "F"
}
