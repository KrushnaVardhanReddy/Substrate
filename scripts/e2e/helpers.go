package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SetupP17Database seeds the required database tables for the phase 17 E2E tests
func SetupP17Database(pool *pgxpool.Pool) error {
	ctx := context.Background()
	_, err := pool.Exec(ctx, "INSERT INTO organizations (github_installation_id, github_org_name) VALUES (12345, 'mgr-org') ON CONFLICT DO NOTHING")
	if err != nil { fmt.Println("org insert failed:", err); return err }
	var orgID string
	err = pool.QueryRow(ctx, "SELECT id FROM organizations WHERE github_org_name = 'mgr-org' LIMIT 1").Scan(&orgID)
	if err != nil { fmt.Println("org query failed:", err); return err }
	_, err = pool.Exec(ctx, "INSERT INTO repositories (id, org_id, github_repo_id, name, full_name) VALUES (gen_random_uuid(), $1, 67890, 'payments-api', 'mgr-org/payments-api') ON CONFLICT DO NOTHING", orgID)
	if err != nil { fmt.Println("repo insert failed:", err); return err }
	_, err = pool.Exec(ctx, "INSERT INTO diff_reports (id, org_name, repo_name, report_data, created_at) VALUES (gen_random_uuid(), 'mgr-org', 'payments-api', '{\"has_breaking_change\": true}', NOW() - INTERVAL '5 days')")
	if err != nil { fmt.Println("diff_report insert failed:", err); return err }
    // Event creation for UI timeline
    _, err = pool.Exec(ctx, "INSERT INTO ecosystem_events (org, repo, event_type, description, event_time) VALUES ('mgr-org', 'payments-api', 'deployment', 'deployed v1.0.0', NOW())")
    if err != nil {
        fmt.Println("Ecosystem events insert failed:", err)
    }
	return err
}
