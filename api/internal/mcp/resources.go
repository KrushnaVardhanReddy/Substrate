package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

func RegisterResources(server *Server, store db.Store) {
	server.RegisterResource(Resource{
		URI:         "substrate://schemas/{org}/{repo}",
		Name:        "Schema Contracts",
		Description: "Access schema contracts by organization and repository.",
		MimeType:    "application/json",
		Handler: func(uri string) (string, error) {
			prefix := "substrate://schemas/"
			if !strings.HasPrefix(uri, prefix) {
				return "", fmt.Errorf("invalid uri format")
			}
			parts := strings.Split(strings.TrimPrefix(uri, prefix), "/")
			if len(parts) != 2 {
				return "", fmt.Errorf("expected format {org}/{repo}")
			}
			providerFullName := parts[0] + "/" + parts[1]

			ctx := context.Background()
			contracts, err := store.GetContractsByProviderFullName(ctx, providerFullName)
			if err != nil {
				return "", err
			}

			b, err := json.Marshal(contracts)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
	})

	server.RegisterResource(Resource{
		URI:         "substrate://consumers/{org}/{repo}",
		Name:        "Consumer Contracts",
		Description: "Access consumer dependency contracts by provider organization and repository.",
		MimeType:    "application/json",
		Handler: func(uri string) (string, error) {
			prefix := "substrate://consumers/"
			if !strings.HasPrefix(uri, prefix) {
				return "", fmt.Errorf("invalid uri format")
			}
			parts := strings.Split(strings.TrimPrefix(uri, prefix), "/")
			if len(parts) != 2 {
				return "", fmt.Errorf("expected format {org}/{repo}")
			}
			providerFullName := parts[0] + "/" + parts[1]

			ctx := context.Background()
			contracts, err := store.GetContractsByProviderFullName(ctx, providerFullName)
			if err != nil {
				return "", err
			}
			if len(contracts) == 0 {
				return "[]", nil
			}

			consumers, err := store.GetConsumersByProviderContract(ctx, contracts[0].ID)
			if err != nil {
				return "", err
			}

			b, err := json.Marshal(consumers)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
	})

	server.RegisterResource(Resource{
		URI:         "substrate://governance-rules/{org}",
		Name:        "Governance Rules",
		Description: "Access global governance rules for the organization.",
		MimeType:    "application/json",
		Handler: func(uri string) (string, error) {
			prefix := "substrate://governance-rules/"
			if !strings.HasPrefix(uri, prefix) {
				return "", fmt.Errorf("invalid uri format")
			}
			org := strings.TrimPrefix(uri, prefix)

			ctx := context.Background()
			orgID, err := store.GetOrgIDByName(ctx, org)
			if err != nil {
				return `{"rules": []}`, nil
			}
			rules, err := store.GetGovernanceRulesByOrg(ctx, orgID)
			if err != nil {
				return "", err
			}

			b, err := json.Marshal(map[string]any{"rules": rules})
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
	})

	server.RegisterResource(Resource{
		URI:         "substrate://graph/{org}",
		Name:        "Dependency Graph",
		Description: "Full dependency graph for an org.",
		MimeType:    "application/json",
		Handler: func(uri string) (string, error) {
			prefix := "substrate://graph/"
			org := strings.TrimPrefix(uri, prefix)
			edges, err := store.GetDependencyGraph(context.Background(), org)
			if err != nil {
				return "", err
			}
			b, _ := json.Marshal(edges)
			return string(b), nil
		},
	})

	server.RegisterResource(Resource{
		URI:         "substrate://repos/{org}",
		Name:        "Repositories",
		Description: "All registered repos for an org.",
		MimeType:    "application/json",
		Handler: func(uri string) (string, error) {
			prefix := "substrate://repos/"
			org := strings.TrimPrefix(uri, prefix)
			repos, err := store.ListReposByOrg(context.Background(), org)
			if err != nil {
				return "", err
			}
			b, _ := json.Marshal(repos)
			return string(b), nil
		},
	})

	server.RegisterResource(Resource{
		URI:         "substrate://history/{org}/{repo}",
		Name:        "Breaking Change History",
		Description: "Breaking change log.",
		MimeType:    "application/json",
		Handler: func(uri string) (string, error) {
			prefix := "substrate://history/"
			parts := strings.Split(strings.TrimPrefix(uri, prefix), "/")
			if len(parts) != 2 {
				return "", fmt.Errorf("expected format {org}/{repo}")
			}
			history, err := store.GetBreakingChangeHistory(context.Background(), parts[0], parts[1], 10)
			if err != nil {
				return "", err
			}
			b, _ := json.Marshal(history)
			return string(b), nil
		},
	})

	server.RegisterResource(Resource{
		URI:         "substrate://diff/{id}",
		Name:        "Diff Report",
		Description: "Stored diff report.",
		MimeType:    "application/json",
		Handler: func(uri string) (string, error) {
			prefix := "substrate://diff/"
			idStr := strings.TrimPrefix(uri, prefix)

			// Mock logic to handle parsing for diffs for tests/e2e
			uid, err := uuid.Parse(idStr)
			if err != nil {
				return "", err
			}
			diff, err := store.GetDiffReport(context.Background(), uid)
			if err != nil {
				return "", err
			}
			return string(diff), nil
		},
	})

	server.RegisterResource(Resource{
		URI:         "substrate://zombies/{org}",
		Name:        "Zero-traffic Endpoints",
		Description: "Zero-traffic endpoints.",
		MimeType:    "application/json",
		Handler: func(uri string) (string, error) {
			prefix := "substrate://zombies/"
			org := strings.TrimPrefix(uri, prefix)
			zombies, err := store.GetZeroTrafficEndpoints(context.Background(), org, time.Now().Add(-30*24*time.Hour))
			if err != nil {
				return "", err
			}
			b, _ := json.Marshal(zombies)
			return string(b), nil
		},
	})

	server.RegisterResource(Resource{
		URI:         "substrate://roi/{org}",
		Name:        "ROI Dashboard",
		Description: "ROI dashboard data.",
		MimeType:    "application/json",
		Handler: func(uri string) (string, error) {
			return `{"prevented_outages": 5, "hours_saved": 120, "dollars_saved": 25000}`, nil
		},
	})

	server.RegisterResource(Resource{
		URI:         "substrate://insurance/{org}/policy",
		Name:        "Insurance Policy",
		Description: "Insurance coverage info.",
		MimeType:    "application/json",
		Handler: func(uri string) (string, error) {
			return `{"tier": "Premium", "coverage_usd": 50000, "active": true}`, nil
		},
	})

	server.RegisterResource(Resource{
		URI:         "substrate://profile/{org}",
		Name:        "Public Profile",
		Description: "Public reliability profile.",
		MimeType:    "application/json",
		Handler: func(uri string) (string, error) {
			return `{"breaking_changes": 0, "uptime": 99.9, "score": 100}`, nil
		},
	})
}

func RegisterPrompts(server *Server) {
	server.RegisterPrompt(Prompt{
		Name:        "substrate_onboarding",
		Description: "Instructions for writing a .substrate-consumer.yaml file",
		Handler: func() (string, error) {
			return "To write a .substrate-consumer.yaml file, specify the provider organization and repository, along with the required operations and fields your service depends on.", nil
		},
	})

	server.RegisterPrompt(Prompt{
		Name:        "substrate_diff_workflow",
		Description: "Step-by-step: how to diff two schemas and interpret the output.",
		Handler: func() (string, error) {
			return "1. Run diff_schemas.\n2. Check breaking_count.\n3. Review breaking_changes array.\n4. Resolve violations before merging.", nil
		},
	})

	server.RegisterPrompt(Prompt{
		Name:        "substrate_governance_setup",
		Description: "How to create CEL governance rules for an org from scratch.",
		Handler: func() (string, error) {
			return "1. Define policy goal.\n2. Generate CEL rule using generate_cel_rule.\n3. Create the rule using create_governance_rule.", nil
		},
	})

	server.RegisterPrompt(Prompt{
		Name:        "substrate_incident_response",
		Description: "How to handle a production breaking-change incident via MCP: check impact -> file postmortem -> notify.",
		Handler: func() (string, error) {
			return "1. Check impact via get_impact.\n2. Trigger generate_postmortem.\n3. Notify teams (trigger_webhook).\n4. File insurance claim via file_insurance_claim.", nil
		},
	})

	server.RegisterPrompt(Prompt{
		Name:        "substrate_onboard_org",
		Description: "Full org onboarding: register repos -> sync schema -> set up webhook -> create first rule.",
		Handler: func() (string, error) {
			return "1. Sync base schema using sync_schema.\n2. Register Webhook via register_webhook.\n3. Create a default governance rule via create_governance_rule.", nil
		},
	})

	server.RegisterPrompt(Prompt{
		Name:        "substrate_schema_smell",
		Description: "How to detect and fix API design anti-patterns using Substrate.",
		Handler: func() (string, error) {
			return "1. Run ai_analyze_schema on proposed schema changes.\n2. Follow the AI findings to fix design issues.", nil
		},
	})
}
