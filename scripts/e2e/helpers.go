package main

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupP17Database seeds the required database tables for the phase 17 E2E tests
func SetupP17Database(pool *pgxpool.Pool) error {
	ctx := context.Background()
	_, err := pool.Exec(ctx, "INSERT INTO organizations (github_org_name) VALUES ('mgr-org') ON CONFLICT DO NOTHING")
	if err != nil { return err }
	var orgID string
	err = pool.QueryRow(ctx, "SELECT id FROM organizations WHERE github_org_name = 'mgr-org' LIMIT 1").Scan(&orgID)
	if err != nil { return err }
	_, err = pool.Exec(ctx, "INSERT INTO repos (id, org_id, name, full_name, is_active) VALUES (gen_random_uuid(), $1, 'payments-api', 'mgr-org/payments-api', true) ON CONFLICT DO NOTHING", orgID)
	if err != nil { return err }
	_, err = pool.Exec(ctx, "INSERT INTO diff_reports (id, org_id, repo_id, has_breaking_change, created_at) VALUES (gen_random_uuid(), $1, (SELECT id FROM repos WHERE name = 'payments-api' LIMIT 1), true, NOW() - INTERVAL '5 days')", orgID)
	if err != nil { return err }
    // Event creation for UI timeline
    _, err = pool.Exec(ctx, "INSERT INTO ecosystem_events (org_id, repo_id, event_type, description) VALUES ($1, (SELECT id FROM repos WHERE name = 'payments-api' LIMIT 1), 'deployment', 'deployed v1.0.0')", orgID)
	return err
}
