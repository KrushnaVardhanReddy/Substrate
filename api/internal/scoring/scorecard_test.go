package scoring

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
	"github.com/jackc/pgx/v5/pgtype"
)

type mockSummary struct {
	BreakingCount int `json:"breaking_count"`
}

type mockDiffReport struct {
	Summary mockSummary `json:"summary"`
}

func TestComputeScore(t *testing.T) {
	ctx := context.Background()
	org := "testorg"
	repo := "testrepo"

	mockStore := &db.MockStore{
		GetScorecardCacheFunc: func(ctx context.Context, arg sqlcgen.GetScorecardCacheParams) (sqlcgen.ScorecardCache, error) {
			// always return empty/error to force computation
			return sqlcgen.ScorecardCache{}, db.ErrNotFound
		},
		UpsertScorecardCacheFunc: func(ctx context.Context, arg sqlcgen.UpsertScorecardCacheParams) error {
			return nil
		},
	}

	tests := []struct {
		name               string
		hasGuides          bool
		soc2Compliant      bool
		hasPii             bool
		breakingChangesNum int
		coveragePercentage float64
		expectedGrade      string
		expectedTotal      int
	}{
		{
			name:               "Grade A (Perfect Score)",
			hasGuides:          true,
			soc2Compliant:      true,
			hasPii:             false,
			breakingChangesNum: 0,
			coveragePercentage: 100.0,
			expectedGrade:      "A",
			expectedTotal:      100, // 25 + 25 + 25 + 25
		},
		{
			name:               "Grade B (Good Score)",
			hasGuides:          true,
			soc2Compliant:      true,
			hasPii:             true, // -5 pts
			breakingChangesNum: 1, // -3 pts
			coveragePercentage: 80.0, // 20 pts
			expectedGrade:      "B",
			expectedTotal:      87, // 25 + 20 + 22 + 20
		},
		{
			name:               "Grade C (Mediocre Score)",
			hasGuides:          false, // 0 pts
			soc2Compliant:      true, // 25 pts
			hasPii:             false,
			breakingChangesNum: 2, // -6 pts -> 19
			coveragePercentage: 64.0, // 16 pts
			expectedGrade:      "C",
			expectedTotal:      60, // 0 + 25 + 19 + 16
		},
		{
			name:               "Grade D (Poor Score)",
			hasGuides:          false, // 0 pts
			soc2Compliant:      false, // -5 pts -> 20
			hasPii:             true, // -5 pts -> 15
			breakingChangesNum: 5, // -15 pts -> 10
			coveragePercentage: 60.0, // 15 pts
			expectedGrade:      "D",
			expectedTotal:      40, // 0 + 15 + 10 + 15
		},
		{
			name:               "Grade F (Failing Score)",
			hasGuides:          false,
			soc2Compliant:      false,
			hasPii:             true,
			breakingChangesNum: 10, // -30 -> 0 pts
			coveragePercentage: 0.0, // 0 pts
			expectedGrade:      "F",
			expectedTotal:      15, // 0 + 15 + 0 + 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock responses
			mockStore.HasRepoGuidesFunc = func(ctx context.Context, arg sqlcgen.HasRepoGuidesParams) (bool, error) {
				return tt.hasGuides, nil
			}

			mockStore.GetRepoComplianceFunc = func(ctx context.Context, arg sqlcgen.GetRepoComplianceParams) (sqlcgen.RepoCompliance, error) {
				return sqlcgen.RepoCompliance{
					Soc2Compliant: tt.soc2Compliant,
					HasPii:        tt.hasPii,
				}, nil
			}

			mockStore.GetLatestQACoverageScoreFunc = func(ctx context.Context, arg sqlcgen.GetLatestQACoverageScoreParams) (pgtype.Numeric, error) {
				var n pgtype.Numeric
				n.Scan(string("100")) // just to format correctly

				// Re-override properly for numeric values
				if tt.coveragePercentage == 100.0 {
					n.Scan("100")
				} else if tt.coveragePercentage == 80.0 {
					n.Scan("80")
				} else if tt.coveragePercentage == 64.0 {
					n.Scan("64")
				} else if tt.coveragePercentage == 60.0 {
					n.Scan("60")
				} else {
					n.Scan("0")
				}

				return n, nil
			}

			// Mock Diff Reports
			var diffs []db.DiffReportRecord
			if tt.breakingChangesNum > 0 {
				rep := mockDiffReport{
					Summary: mockSummary{
						BreakingCount: tt.breakingChangesNum,
					},
				}
				repJSON, _ := json.Marshal(rep)
				diffs = append(diffs, db.DiffReportRecord{
					ReportData: repJSON,
					CreatedAt:  time.Now(),
				})
			}
			mockStore.GetDiffReportsByRepoFunc = func(ctx context.Context, orgName, repoName string, limit int) ([]db.DiffReportRecord, error) {
				return diffs, nil
			}

			scorecard, err := ComputeScore(ctx, mockStore, org, repo)
			if err != nil {
				t.Fatalf("ComputeScore returned error: %v", err)
			}

			if scorecard.Grade != tt.expectedGrade {
				t.Errorf("Expected grade %s, got %s", tt.expectedGrade, scorecard.Grade)
			}
			if scorecard.Total != tt.expectedTotal {
				t.Errorf("Expected total %d, got %d", tt.expectedTotal, scorecard.Total)
			}
		})
	}
}
