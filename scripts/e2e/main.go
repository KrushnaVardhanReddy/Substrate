package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"e2e/scenarios/graphql"
	"e2e/scenarios/openapi"
	"e2e/scenarios/sql"

	"github.com/google/go-github/v62/github"
)

func main() {
	scenario := flag.String("scenario", "all", "The scenario to run (e.g., all, openapi-breaking, sql-breaking, etc)")
	flag.Parse()

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("GITHUB_TOKEN is required")
	}

	client := github.NewClient(nil).WithAuthToken(token)
	ctx := context.Background()

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		log.Fatalf("Failed to get user: %v", err)
	}
	owner := user.GetLogin()
	log.Printf("Running E2E tests as %s", owner)

	switch strings.ToLower(*scenario) {
	case "all":
		openapi.RunAll(ctx, client, owner)
		sql.RunAll(ctx, client, owner)
		graphql.RunAll(ctx, client, owner)
	case "openapi":
		openapi.RunAll(ctx, client, owner)
	case "sql":
		sql.RunAll(ctx, client, owner)
	case "graphql":
		graphql.RunAll(ctx, client, owner)
	case "openapi-breaking":
		openapi.RunBreaking(ctx, client, owner)
	case "openapi-safe":
		openapi.RunSafe(ctx, client, owner)
	case "openapi-override":
		openapi.RunOverride(ctx, client, owner)
	case "openapi-warning":
		openapi.RunWarning(ctx, client, owner)
	case "sql-breaking":
		sql.RunBreaking(ctx, client, owner)
	case "sql-safe":
		sql.RunSafe(ctx, client, owner)
	case "sql-override":
		sql.RunOverride(ctx, client, owner)
	case "sql-warning":
		sql.RunWarning(ctx, client, owner)
	case "graphql-breaking":
		graphql.RunBreaking(ctx, client, owner)
	case "graphql-safe":
		graphql.RunSafe(ctx, client, owner)
	case "graphql-override":
		graphql.RunOverride(ctx, client, owner)
	case "graphql-warning":
		graphql.RunWarning(ctx, client, owner)
	default:
		log.Fatalf("Unknown scenario: %s.", *scenario)
	}
}
