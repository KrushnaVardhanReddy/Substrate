package main

import (
	"context"
	"flag"
	"log"
	"os"
	"strings"

	"e2e/scenarios/aiml"
	"e2e/scenarios/asyncapi"
	"e2e/scenarios/avro"
	"e2e/scenarios/graphql"
	"e2e/scenarios/openapi"
	"e2e/scenarios/protobuf"
	"e2e/scenarios/sql"
	"e2e/scenarios/terraform"

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
		protobuf.RunAll(ctx, client, owner)
		asyncapi.RunAll(ctx, client, owner)
		avro.RunAll(ctx, client, owner)
		terraform.RunAll(ctx, client, owner)
		aiml.RunAll(ctx, client, owner)
	case "openapi":
		openapi.RunAll(ctx, client, owner)
	case "sql":
		sql.RunAll(ctx, client, owner)
	case "graphql":
		graphql.RunAll(ctx, client, owner)
	case "protobuf":
		protobuf.RunAll(ctx, client, owner)
	case "asyncapi":
		asyncapi.RunAll(ctx, client, owner)
	case "asyncapi-breaking":
		asyncapi.RunBreaking(ctx, client, owner)
	case "asyncapi-safe":
		asyncapi.RunSafe(ctx, client, owner)
	case "asyncapi-override":
		asyncapi.RunOverride(ctx, client, owner)
	case "asyncapi-warning":
		asyncapi.RunWarning(ctx, client, owner)
	case "avro":
		avro.RunAll(ctx, client, owner)
	case "avro-breaking":
		avro.RunBreaking(ctx, client, owner)
	case "avro-safe":
		avro.RunSafe(ctx, client, owner)
	case "avro-override":
		avro.RunOverride(ctx, client, owner)
	case "terraform":
		terraform.RunAll(ctx, client, owner)
	case "terraform-breaking":
		terraform.RunBreaking(ctx, client, owner)
	case "terraform-safe":
		terraform.RunSafe(ctx, client, owner)
	case "terraform-override":
		terraform.RunOverride(ctx, client, owner)
	case "aiml":
		aiml.RunAll(ctx, client, owner)
	case "aiml-breaking":
		aiml.RunBreaking(ctx, client, owner)
	case "aiml-safe":
		aiml.RunSafe(ctx, client, owner)
	case "aiml-override":
		aiml.RunOverride(ctx, client, owner)
	case "openapi-breaking":
		openapi.RunBreaking(ctx, client, owner)
	case "openapi-safe":
		openapi.RunSafe(ctx, client, owner)
	case "openapi-override":
		openapi.RunOverride(ctx, client, owner)
	case "openapi-warning":
		openapi.RunWarning(ctx, client, owner)
	case "openapi-pii":
		openapi.RunPIIAuditing(ctx, client, owner)
	case "openapi-traffic":
		openapi.RunTrafficAware(ctx, client, owner)
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
	case "protobuf-breaking":
		protobuf.RunBreaking(ctx, client, owner)
	case "protobuf-safe":
		protobuf.RunSafe(ctx, client, owner)
	case "protobuf-override":
		protobuf.RunOverride(ctx, client, owner)
	case "protobuf-warning":
		protobuf.RunWarning(ctx, client, owner)
	default:
		log.Fatalf("Unknown scenario: %s.", *scenario)
	}
}
