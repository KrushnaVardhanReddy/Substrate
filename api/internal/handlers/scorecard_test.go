package handlers

import (
	"encoding/json"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestGetScorecardHandler(t *testing.T) {
	mockStore := &db.MockStore{
		GetScorecardCacheFunc: func(ctx context.Context, arg sqlcgen.GetScorecardCacheParams) (sqlcgen.ScorecardCache, error) {
			return sqlcgen.ScorecardCache{}, db.ErrNotFound
		},
		UpsertScorecardCacheFunc: func(ctx context.Context, arg sqlcgen.UpsertScorecardCacheParams) error {
			return nil
		},
		HasRepoGuidesFunc: func(ctx context.Context, arg sqlcgen.HasRepoGuidesParams) (bool, error) {
			return true, nil // 25
		},
		GetRepoComplianceFunc: func(ctx context.Context, arg sqlcgen.GetRepoComplianceParams) (sqlcgen.RepoCompliance, error) {
			return sqlcgen.RepoCompliance{Soc2Compliant: true, HasPii: false}, nil // 25
		},
		GetLatestQACoverageScoreFunc: func(ctx context.Context, arg sqlcgen.GetLatestQACoverageScoreParams) (pgtype.Numeric, error) {
			var n pgtype.Numeric
			n.Scan("100")
			n.Valid = true
			return n, nil // 25
		},
		GetDiffReportsByRepoFunc: func(ctx context.Context, org, repo string, limit int) ([]db.DiffReportRecord, error) {
			return []db.DiffReportRecord{}, nil // 25
		},
	}

	req, _ := http.NewRequest("GET", "/api/v1/scorecard/mcp-org/core-repo", nil)
	req.SetPathValue("org", "mcp-org")
	req.SetPathValue("repo", "core-repo")
	w := httptest.NewRecorder()

	GetScorecardHandler(mockStore)(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var respMap map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &respMap)
	delete(respMap, "computed_at")
	expMap := map[string]interface{}{
		"grade": "A",
		"total": float64(100),
		"breakdown": map[string]interface{}{
			"docs": float64(25),
			"compliance": float64(25),
			"volatility": float64(25),
			"coverage": float64(25),
		},
	}
	assert.Equal(t, expMap, respMap)
}
