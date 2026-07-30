//go:build ignore
// +build ignore

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable"
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Println("Failed to connect to db:", err)
		os.Exit(1)
	}
	defer pool.Close()

	_, err = pool.Exec(ctx, "INSERT INTO organizations (github_installation_id, github_org_name) VALUES (916, 'e2e-org') ON CONFLICT DO NOTHING")
	if err != nil { fmt.Println("org insert failed:", err); os.Exit(1) }

	var orgID string
	err = pool.QueryRow(ctx, "SELECT id FROM organizations WHERE github_org_name = 'e2e-org' LIMIT 1").Scan(&orgID)
	if err != nil { fmt.Println("org query failed:", err); os.Exit(1) }

	_, err = pool.Exec(ctx, "INSERT INTO repositories (id, org_id, github_repo_id, name, full_name, base_url) VALUES (gen_random_uuid(), $1, 916916, 'guide-api-repo', 'e2e-org/guide-api-repo', 'http://localhost:8081') ON CONFLICT DO NOTHING", orgID)
	if err != nil { fmt.Println("repo insert failed:", err); os.Exit(1) }

	_, err = pool.Exec(ctx, "INSERT INTO repo_guides (org, repo, file_path, title, content) VALUES ('e2e-org', 'guide-api-repo', 'docs/authentication.md', 'Authentication Guide', '# Auth\nUse Bearer tokens.') ON CONFLICT DO NOTHING")
	if err != nil { fmt.Println("repo_guides insert failed:", err); os.Exit(1) }

	fmt.Println("P16 seeded successfully")
}
