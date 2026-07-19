package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (s *PGStore) UpsertRepoMetric(ctx context.Context, repoID uuid.UUID, score int) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO repo_metrics (repo_id, predictive_risk_score, calculated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (repo_id) DO UPDATE SET
			predictive_risk_score = EXCLUDED.predictive_risk_score,
			calculated_at = NOW()
	`, repoID, score)
	return err
}

func (s *PGStore) GetAllRepositories(ctx context.Context) ([]Repository, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, org_id, github_repo_id, name, full_name, metadata, created_at
		FROM repositories
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []Repository
	for rows.Next() {
		var r Repository
		err := rows.Scan(&r.ID, &r.OrgID, &r.GithubRepoID, &r.Name, &r.FullName, &r.Metadata, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		repos = append(repos, r)
	}
	return repos, rows.Err()
}

func (s *PGStore) CountRecentBreakingChanges(ctx context.Context, repoID uuid.UUID, since time.Time) (int, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM breaking_change_records
		WHERE repo_id = $1 AND timestamp >= $2
	`, repoID, since).Scan(&count)
	return count, err
}

func (s *PGStore) GetCommitVelocity(ctx context.Context, repoID uuid.UUID, since time.Time) (int, error) {
	// For now, we don't have a granular commit table.
	// Returning a mock value or 0 for now.
	// In the future, this would query a commits table or contracts table.
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM contracts
		WHERE repo_id = $1 AND synced_at >= $2
	`, repoID, since).Scan(&count)
	return count, err
}

func (s *PGStore) GetTimeSinceLastBreak(ctx context.Context, repoID uuid.UUID) (*time.Time, error) {
	var lastBreak time.Time
	err := s.pool.QueryRow(ctx, `
		SELECT MAX(timestamp)
		FROM breaking_change_records
		WHERE repo_id = $1
	`).Scan(&lastBreak)

	if err != nil {
		// If no rows or null max
		if err.Error() == "sql: Scan error on column index 0, name \"max\": converting NULL to time.Time is unsupported" || err.Error() == "no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return &lastBreak, nil
}
