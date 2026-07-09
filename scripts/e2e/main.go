package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"e2e/scenarios/openapi"
	// "e2e/scenarios/sql"
	// "e2e/scenarios/graphql"

	"github.com/google/go-github/v62/github"
)

func main() {
	scenario := flag.String("scenario", "all", "The scenario to run (e.g., all, openapi-breaking, openapi-safe, openapi-override, openapi-warning)")
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
	case "openapi-breaking":
		openapi.RunBreaking(ctx, client, owner)
	case "openapi-safe":
		openapi.RunSafe(ctx, client, owner)
	case "openapi-override":
		openapi.RunOverride(ctx, client, owner)
	case "openapi-warning":
		openapi.RunWarning(ctx, client, owner)
	default:
		log.Fatalf("Unknown scenario: %s. Available options: all, openapi-breaking, openapi-safe, openapi-override, openapi-warning", *scenario)
	}
}
