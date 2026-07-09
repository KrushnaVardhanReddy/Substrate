package main

import (
	"context"
	"log"
	"os"

	"e2e/scenarios/openapi"
	// "e2e/scenarios/sql"
	// "e2e/scenarios/graphql"

	"github.com/google/go-github/v62/github"
)

func main() {
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

	// We can easily run each scenario here
	openapi.Run(ctx, client, owner)
	// sql.Run(ctx, client, owner)
	// graphql.Run(ctx, client, owner)
}
