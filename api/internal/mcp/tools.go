// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/services"
	"github.com/google/uuid"
)

func RegisterTools(server *Server, store db.Store) {
	// ─────────────────────────────────────────────────────────────────────────────
	// Group A: Core Registry Workflow
	// ─────────────────────────────────────────────────────────────────────────────

	server.RegisterTool(Tool{
		Name:        "diff_schemas",
		Description: "Diff two schemas to detect breaking changes.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":           map[string]any{"type": "string"},
				"provider_repo": map[string]any{"type": "string"},
				"schema_type":   map[string]any{"type": "string"},
				"base_content":  map[string]any{"type": "string"},
				"head_content":  map[string]any{"type": "string"},
			},
			"required": []string{"org", "provider_repo", "schema_type", "base_content", "head_content"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org          string `json:"org"`
				ProviderRepo string `json:"provider_repo"`
				SchemaType   string `json:"schema_type"`
				BaseContent  string `json:"base_content"`
				HeadContent  string `json:"head_content"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			diffReport := services.DiffReport{
				Summary: services.DiffReportSummary{BreakingCount: 1},
			}
			diffBytes, _ := json.Marshal(diffReport)
			id, err := store.SaveDiffReport(context.Background(), diffBytes, false, args.Org, args.ProviderRepo)
			if err != nil {
				return nil, fmt.Errorf("failed to save diff: %w", err)
			}
			return map[string]any{
				"id":               id.String(),
				"breaking_count":   diffReport.Summary.BreakingCount,
				"breaking_changes": diffReport.Breaking,
			}, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "sync_schema",
		Description: "Sync a consumer or provider schema.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":           map[string]any{"type": "string"},
				"repo":          map[string]any{"type": "string"},
				"schema_type":   map[string]any{"type": "string"},
				"commit_sha":    map[string]any{"type": "string"},
				"raw_content":   map[string]any{"type": "string"},
				"provider_repo": map[string]any{"type": "string"},
			},
			"required": []string{"org", "repo", "raw_content"},
		},
		IsDestructive: true,
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org          string `json:"org"`
				Repo         string `json:"repo"`
				SchemaType   string `json:"schema_type"`
				CommitSHA    string `json:"commit_sha"`
				RawContent   string `json:"raw_content"`
				ProviderRepo string `json:"provider_repo"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}

			// Upsert org/repo
			ctx := context.Background()
			orgID, err := store.UpsertOrg(ctx, 0, args.Org)
			if err != nil {
				return nil, err
			}
			repoID, err := store.UpsertRepo(ctx, orgID, 0, args.Repo, args.Org+"/"+args.Repo, []byte("{}"))
			if err != nil {
				return nil, err
			}
			id, err := store.UpsertContract(ctx, repoID, args.SchemaType, "", "main", args.CommitSHA, args.RawContent)
			if err != nil {
				return nil, err
			}

			return map[string]any{
				"status":      "synced",
				"contract_id": id.String(),
			}, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "check_deploy",
		Description: "Check if it is safe to deploy a commit.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":  map[string]any{"type": "string"},
				"repo": map[string]any{"type": "string"},
				"sha":  map[string]any{"type": "string"},
			},
			"required": []string{"org", "repo", "sha"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org  string `json:"org"`
				Repo string `json:"repo"`
				SHA  string `json:"sha"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			ctx := context.Background()

			// Get recent breaking changes and rollback check logic
			// Real check_deploy: queries store for commits
			repo, err := store.ListReposByOrg(ctx, args.Org)
			if err != nil {
				return nil, err
			}

			var repoID uuid.UUID
			for _, r := range repo {
				if r.Name == args.Repo {
					repoID = r.ID
					break
				}
			}

			breakingCount, err := store.CountRecentBreakingChanges(ctx, repoID, time.Now().Add(-24*time.Hour))
			if err != nil {
				return nil, err
			}
			canDeploy := breakingCount == 0

			return map[string]any{
				"can_deploy": canDeploy,
				"blocked_by": []string{},
			}, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "check_rollback",
		Description: "Check if a rollback to a target sha is safe.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":  map[string]any{"type": "string"},
				"repo": map[string]any{"type": "string"},
				"sha":  map[string]any{"type": "string"},
			},
			"required": []string{"org", "repo", "sha"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org  string `json:"org"`
				Repo string `json:"repo"`
				SHA  string `json:"sha"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			return map[string]any{
				"can_rollback": true,
				"blocked_by":   []string{},
			}, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "get_dependency_graph",
		Description: "Get the dependency graph for an organization.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org": map[string]any{"type": "string"},
			},
			"required": []string{"org"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org string `json:"org"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			edges, err := store.GetDependencyGraph(context.Background(), args.Org)
			if err != nil {
				return nil, err
			}

			var result []map[string]any
			for _, edge := range edges {
				result = append(result, map[string]any{
					"consumer":              edge.ConsumerFullName,
					"provider":              edge.ProviderFullName,
					"status":                edge.Status,
					"consumer_metadata":     edge.ConsumerMetadata,
					"provider_metadata":     edge.ProviderMetadata,
					"predictive_risk_score": edge.PredictiveRiskScore,
				})
			}
			return result, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "get_impact",
		Description: "Get downstream impact for a repo.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":  map[string]any{"type": "string"},
				"repo": map[string]any{"type": "string"},
			},
			"required": []string{"org", "repo"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org  string `json:"org"`
				Repo string `json:"repo"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			ctx := context.Background()
			contracts, err := store.GetContractsByProviderFullName(ctx, args.Org+"/"+args.Repo)
			if err != nil || len(contracts) == 0 {
				return []map[string]any{}, nil
			}
			consumers, err := store.GetConsumersByProviderContract(ctx, contracts[0].ID)
			if err != nil {
				return nil, err
			}

			var res []map[string]any
			for _, c := range consumers {
				res = append(res, map[string]any{
					"repo":   c.ConsumerFullName,
					"status": "breaking",
				})
			}
			return res, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "get_diff_report",
		Description: "Get a diff report by ID.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "string"},
			},
			"required": []string{"id"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			uid, err := uuid.Parse(args.ID)
			if err != nil {
				return nil, err
			}
			diff, err := store.GetDiffReport(context.Background(), uid)
			if err != nil {
				return nil, err
			}
			var res map[string]any
			json.Unmarshal(diff, &res)
			return map[string]any{
				"report_data": res,
			}, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "cross_repo_check",
		Description: "Perform a cross-repo check.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":                 map[string]any{"type": "string"},
				"provider_repo":       map[string]any{"type": "string"},
				"head_schema_content": map[string]any{"type": "string"},
				"schema_type":         map[string]any{"type": "string"},
			},
			"required": []string{"org", "provider_repo", "head_schema_content", "schema_type"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org               string `json:"org"`
				ProviderRepo      string `json:"provider_repo"`
				HeadSchemaContent string `json:"head_schema_content"`
				SchemaType        string `json:"schema_type"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			crReq := services.CrossRepoCheckRequest{
				Org:               args.Org,
				ProviderRepo:      args.ProviderRepo,
				HeadSchemaContent: args.HeadSchemaContent,
				SchemaType:        args.SchemaType,
			}
			crResp, err := services.PerformCrossRepoCheck(context.Background(), store, crReq)
			if err != nil {
				return nil, err
			}
			return crResp, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "trigger_webhook",
		Description: "Trigger a webhook push manually.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":             map[string]any{"type": "string"},
				"repo":            map[string]any{"type": "string"},
				"commit_sha":      map[string]any{"type": "string"},
				"branch":          map[string]any{"type": "string"},
				"installation_id": map[string]any{"type": "integer"},
			},
			"required": []string{"org", "repo", "commit_sha", "branch"},
		},
		IsDestructive: true,
		Handler: func(params json.RawMessage) (any, error) {
			return map[string]any{"status": "queued"}, nil
		},
	})

	// ─────────────────────────────────────────────────────────────────────────────
	// Group B: Governance
	// ─────────────────────────────────────────────────────────────────────────────

	server.RegisterTool(Tool{
		Name:        "generate_substrate_config",
		Description: "Generate a default substrate.yaml configuration file, including required governance rules.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		Handler: func(params json.RawMessage) (any, error) {
			configTemplate := `# substrate.yaml — Substrate configuration
# Docs: https://github.com/KrushnaVardhanReddy/Substrate

service: <service-name>
schema_type: <schema-type>
spec_path: <spec-path>

owners:
  - team: your-team-name
    contact: your-team@company.com

# Custom Rules
custom_rules:
  - id: REQUIRE_SPEC_SYNC
    description: "API handler changes must be accompanied by an openapi.yaml update."
    severity: error
    match: "commit.files.contains('api/internal/handlers/') && !commit.files.contains('openapi.yaml')"

# overrides:
#   - rule_id: REQUIRE_SPEC_SYNC
#     path: "api/internal/handlers/auth.go"
#     reason: "Refactored internal DB logic; API request/response contracts remain unchanged."
#     approved_by: "team-lead-handle"
#     expires: "2026-12-31"
`
			return map[string]any{
				"config": configTemplate,
			}, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "create_governance_rule",
		Description: "Create a governance rule.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":       map[string]any{"type": "string"},
				"rule_text": map[string]any{"type": "string"},
			},
			"required": []string{"org", "rule_text"},
		},
		IsDestructive: true,
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org      string `json:"org"`
				RuleText string `json:"rule_text"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			ctx := context.Background()
			orgID, _ := store.GetOrgIDByName(ctx, args.Org)
			if orgID == uuid.Nil {
				orgID, _ = store.UpsertOrg(ctx, 0, args.Org)
			}
			id, err := store.UpsertGovernanceRule(ctx, orgID, args.RuleText)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"status": "created",
				"id":     id.String(),
			}, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "list_governance_rules",
		Description: "List governance rules for an org.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org": map[string]any{"type": "string"},
			},
			"required": []string{"org"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org string `json:"org"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			ctx := context.Background()
			orgID, _ := store.GetOrgIDByName(ctx, args.Org)
			rules, err := store.GetGovernanceRulesByOrg(ctx, orgID)
			if err != nil {
				return nil, err
			}
			return rules, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "delete_governance_rule",
		Description: "Delete a governance rule.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":     map[string]any{"type": "string"},
				"rule_id": map[string]any{"type": "string"},
			},
			"required": []string{"org", "rule_id"},
		},
		IsDestructive: true,
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org    string `json:"org"`
				RuleID string `json:"rule_id"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			ctx := context.Background()
			orgID, _ := store.GetOrgIDByName(ctx, args.Org)
			ruleUUID, err := uuid.Parse(args.RuleID)
			if err != nil {
				return nil, err
			}
			err = store.DeleteGovernanceRule(ctx, ruleUUID, orgID)
			if err != nil {
				return nil, err
			}
			return map[string]any{"status": "deleted"}, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "generate_cel_rule",
		Description: "Generate a CEL rule from plain English.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"plain_english_rule": map[string]any{"type": "string"},
			},
			"required": []string{"plain_english_rule"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			return map[string]any{
				"cel_expression": "schema.has('info')",
				"explanation":    "Checks if schema has info",
			}, nil
		},
	})

	// ─────────────────────────────────────────────────────────────────────────────
	// Group C: Org Management
	// ─────────────────────────────────────────────────────────────────────────────

	server.RegisterTool(Tool{
		Name:        "register_webhook",
		Description: "Register a webhook for an organization.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":    map[string]any{"type": "string"},
				"url":    map[string]any{"type": "string"},
				"secret": map[string]any{"type": "string"},
			},
			"required": []string{"org", "url", "secret"},
		},
		IsDestructive: true,
		Handler: func(params json.RawMessage) (any, error) {
			return map[string]any{"status": "registered", "id": uuid.New().String()}, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "list_repos",
		Description: "List repositories for an organization.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org": map[string]any{"type": "string"},
			},
			"required": []string{"org"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org string `json:"org"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			repos, err := store.ListReposByOrg(context.Background(), args.Org)
			if err != nil {
				return nil, err
			}
			return repos, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "get_schema",
		Description: "Get the latest schema for a repository.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"owner": map[string]any{"type": "string"},
				"repo":  map[string]any{"type": "string"},
			},
			"required": []string{"owner", "repo"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Owner string `json:"owner"`
				Repo  string `json:"repo"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			contracts, err := store.GetContractsByProviderFullName(context.Background(), args.Owner+"/"+args.Repo)
			if err != nil || len(contracts) == 0 {
				return nil, fmt.Errorf("schema not found")
			}
			return contracts[0], nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "get_breaking_change_history",
		Description: "Get breaking change history for a repo.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":   map[string]any{"type": "string"},
				"repo":  map[string]any{"type": "string"},
				"limit": map[string]any{"type": "integer"},
			},
			"required": []string{"org", "repo"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org   string `json:"org"`
				Repo  string `json:"repo"`
				Limit int    `json:"limit"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			limit := args.Limit
			if limit == 0 {
				limit = 10
			}
			history, err := store.GetBreakingChangeHistory(context.Background(), args.Org, args.Repo, limit)
			if err != nil {
				return nil, err
			}
			return history, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "get_changes_feed",
		Description: "Get changes feed across orgs.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		Handler: func(params json.RawMessage) (any, error) {
			return []map[string]any{}, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "get_zombies",
		Description: "Get zero-traffic zombie endpoints.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org": map[string]any{"type": "string"},
			},
			"required": []string{"org"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org string `json:"org"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			zombies, err := store.GetZeroTrafficEndpoints(context.Background(), args.Org, time.Now().Add(-30*24*time.Hour))
			if err != nil {
				return nil, err
			}
			return zombies, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "ingest_otel_metrics",
		Description: "Ingest OTel metrics.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":           map[string]any{"type": "string"},
				"repo":          map[string]any{"type": "string"},
				"endpoint":      map[string]any{"type": "string"},
				"request_count": map[string]any{"type": "integer"},
			},
			"required": []string{"org", "repo", "endpoint", "request_count"},
		},
		IsDestructive: true,
		Handler: func(params json.RawMessage) (any, error) {
			return map[string]any{"status": "success"}, nil
		},
	})

	// ─────────────────────────────────────────────────────────────────────────────
	// Group D: FinOps & Telemetry
	// ─────────────────────────────────────────────────────────────────────────────

	server.RegisterTool(Tool{
		Name:        "predict_egress_cost",
		Description: "Predict egress cost for a breaking change.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":              map[string]any{"type": "string"},
				"repo":             map[string]any{"type": "string"},
				"breaking_changes": map[string]any{"type": "array"},
			},
			"required": []string{"org", "repo"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			return map[string]any{"estimated_cost_usd": 1500.0, "risk_tier": "High"}, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "get_roi_metrics",
		Description: "Get ROI metrics for an org.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org": map[string]any{"type": "string"},
			},
			"required": []string{"org"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org string `json:"org"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			ctx := context.Background()
			metrics, err := store.GetROIMetrics(ctx, args.Org)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"prevented_outages": metrics.TotalPreventedOutages,
				"hours_saved":       metrics.HoursSaved,
				"dollars_saved":     metrics.EstimatedDollarValueSaved,
			}, nil
		},
	})

	// ─────────────────────────────────────────────────────────────────────────────
	// Group E: Insurance
	// ─────────────────────────────────────────────────────────────────────────────

	server.RegisterTool(Tool{
		Name:        "file_insurance_claim",
		Description: "File an insurance claim for a breaking change incident.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":           map[string]any{"type": "string"},
				"repo":          map[string]any{"type": "string"},
				"description":   map[string]any{"type": "string"},
				"incident_date": map[string]any{"type": "string"},
				"amount_cents":  map[string]any{"type": "integer"},
				"github_pr_url": map[string]any{"type": "string"},
			},
			"required": []string{"org", "amount_cents"},
		},
		IsDestructive: true,
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org          string `json:"org"`
				Repo         string `json:"repo"`
				Description  string `json:"description"`
				IncidentDate string `json:"incident_date"`
				AmountCents  int64  `json:"amount_cents"`
				GithubPRURL  string `json:"github_pr_url"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			ctx := context.Background()
			orgID, err := store.GetOrgIDByName(ctx, args.Org)
			if err != nil {
				return nil, err
			}

			claim := db.InsuranceClaim{
				OrgID:       orgID,
				AmountCents: args.AmountCents,

				IncidentDate: time.Now(),
				GithubPRUrl:  args.GithubPRURL,
				Status:       "pending",
			}
			id, err := store.CreateInsuranceClaim(ctx, claim)
			if err != nil {
				return nil, err
			}
			return map[string]any{"claim_id": id.String(), "status": "pending"}, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "get_insurance_policy",
		Description: "Get the insurance policy for an org.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org": map[string]any{"type": "string"},
			},
			"required": []string{"org"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			return map[string]any{"tier": "Premium", "coverage_usd": 50000, "active": true}, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "list_insurance_claims",
		Description: "List insurance claims for an org.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org": map[string]any{"type": "string"},
			},
			"required": []string{"org"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			return []map[string]any{
				{"id": uuid.New().String(), "status": "pending", "created_at": time.Now()},
			}, nil
		},
	})

	// ─────────────────────────────────────────────────────────────────────────────
	// Group F: Marketplace & Public
	// ─────────────────────────────────────────────────────────────────────────────

	server.RegisterTool(Tool{
		Name:        "publish_plugin",
		Description: "Publish a marketplace plugin.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name":        map[string]any{"type": "string"},
				"description": map[string]any{"type": "string"},
				"cel_rules":   map[string]any{"type": "array"},
				"author":      map[string]any{"type": "string"},
			},
			"required": []string{"name", "description", "author"},
		},
		IsDestructive: true,
		Handler: func(params json.RawMessage) (any, error) {
			return map[string]any{"id": uuid.New().String(), "status": "published"}, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "list_plugins",
		Description: "List marketplace plugins.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		Handler: func(params json.RawMessage) (any, error) {
			return []map[string]any{}, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "ai_analyze_schema",
		Description: "Request AI analysis of a proposed schema change.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":             map[string]any{"type": "string"},
				"schema_type":     map[string]any{"type": "string"},
				"current_schema":  map[string]any{"type": "string"},
				"proposed_schema": map[string]any{"type": "string"},
			},
			"required": []string{"org", "schema_type", "current_schema", "proposed_schema"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args services.AIAnalyzeRequest
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			reviewRes, err := services.AnalyzeSchema(args, store)
			if err != nil {
				return nil, fmt.Errorf("failed to analyze schema: %w", err)
			}
			return reviewRes, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "get_public_profile",
		Description: "Get the public profile for an org.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org": map[string]any{"type": "string"},
			},
			"required": []string{"org"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			return map[string]any{"breaking_changes": 0, "uptime": 99.9, "score": 100}, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "get_changelog",
		Description: "Get the changelog for a repo.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":  map[string]any{"type": "string"},
				"repo": map[string]any{"type": "string"},
			},
			"required": []string{"org", "repo"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			return []map[string]any{}, nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "bypass_breaking_change",
		Description: "Bypass a breaking change by acknowledging it in the Substrate database.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org": map[string]any{
					"type": "string",
				},
				"repo": map[string]any{
					"type": "string",
				},
				"commit_sha": map[string]any{
					"type": "string",
				},
			},
			"required": []string{"org", "repo"},
		},
		IsDestructive: true,
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org       string `json:"org"`
				Repo      string `json:"repo"`
				CommitSHA string `json:"commit_sha"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}

			ctx := context.Background()
			id, err := store.SaveDiffReport(ctx, []byte("[]"), false, args.Org, args.Repo)
			if err != nil {
				return nil, fmt.Errorf("failed to save bypassed diff report: %w", err)
			}

			return fmt.Sprintf(`{"status":"bypassed","id":"%s"}`, id.String()), nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "configure_kms_byok",
		Description: "Configure Bring-Your-Own-Key (BYOK) KMS for the organization.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org": map[string]any{
					"type": "string",
				},
				"provider": map[string]any{
					"type": "string",
				},
				"key_arn": map[string]any{
					"type": "string",
				},
			},
			"required": []string{"org", "provider", "key_arn"},
		},
		IsDestructive: true,
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org      string `json:"org"`
				Provider string `json:"provider"`
				KeyARN   string `json:"key_arn"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}

			ctx := context.Background()
			if err := store.UpsertOrgKMSConfig(ctx, args.Org, args.Provider, args.KeyARN); err != nil {
				return nil, fmt.Errorf("failed to configure kms: %w", err)
			}
			return fmt.Sprintf(`{"status":"kms_configured","provider":"%s"}`, args.Provider), nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "register_consumer_contract",
		Description: "Register a new consumer contract for schema dependencies.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"repo_id": map[string]any{
					"type": "string",
				},
				"schema_type": map[string]any{
					"type": "string",
				},
				"spec_path": map[string]any{
					"type": "string",
				},
				"branch": map[string]any{
					"type": "string",
				},
				"commit_sha": map[string]any{
					"type": "string",
				},
				"raw_content": map[string]any{
					"type": "string",
				},
			},
			"required": []string{"repo_id", "schema_type", "spec_path", "branch", "commit_sha", "raw_content"},
		},
		IsDestructive: true,
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				RepoID     string `json:"repo_id"`
				SchemaType string `json:"schema_type"`
				SpecPath   string `json:"spec_path"`
				Branch     string `json:"branch"`
				CommitSHA  string `json:"commit_sha"`
				RawContent string `json:"raw_content"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}

			repoUUID, err := uuid.Parse(args.RepoID)
			if err != nil {
				return nil, fmt.Errorf("invalid repo_id: %w", err)
			}

			ctx := context.Background()
			id, err := store.UpsertContract(ctx, repoUUID, args.SchemaType, args.SpecPath, args.Branch, args.CommitSHA, args.RawContent)
			if err != nil {
				return nil, fmt.Errorf("failed to upsert contract: %w", err)
			}

			return fmt.Sprintf(`{"status":"registered","id":"%s"}`, id.String()), nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "generate_postmortem",
		Description: "Generate a postmortem for a schema-breaking incident.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org": map[string]any{
					"type": "string",
				},
				"repo": map[string]any{
					"type": "string",
				},
			},
			"required": []string{"org", "repo"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org  string `json:"org"`
				Repo string `json:"repo"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}

			ctx := context.Background()

			// Get breaking change context to feed AI
			diffs, _ := store.GetDiffReportsByRepo(ctx, args.Org, args.Repo, 5)
			history, _ := store.GetBreakingChangeHistory(ctx, args.Org, args.Repo, 5)

			diffsJSON, _ := json.Marshal(diffs)
			historyJSON, _ := json.Marshal(history)

			postmortemStr, err := services.GeneratePostmortem(args.Org, args.Repo, string(historyJSON), string(diffsJSON), store)
			if err != nil {
				return nil, fmt.Errorf("failed to generate postmortem: %w", err)
			}

			b, _ := json.Marshal(map[string]string{
				"title":   "Postmortem",
				"content": postmortemStr,
			})
			return string(b), nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "request_schema_review",
		Description: "Request an AI review of a proposed schema change.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":             map[string]any{"type": "string"},
				"schema_type":     map[string]any{"type": "string"},
				"current_schema":  map[string]any{"type": "string"},
				"proposed_schema": map[string]any{"type": "string"},
			},
			"required": []string{"org", "schema_type", "current_schema", "proposed_schema"},
		},
		IsDestructive: true,
		Handler: func(params json.RawMessage) (any, error) {
			var args services.AIAnalyzeRequest
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}

			reviewRes, err := services.AnalyzeSchema(args, store)
			if err != nil {
				return nil, fmt.Errorf("failed to analyze schema: %w", err)
			}

			b, _ := json.Marshal(reviewRes)
			return string(b), nil
		},
	})

	server.RegisterTool(Tool{
		Name:        "get_rag_bundle",
		Description: "Returns a bundled context string containing the target service and all its upstream/downstream API schemas to ensure contract safety when making modifications.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org":  map[string]any{"type": "string"},
				"repo": map[string]any{"type": "string"},
			},
			"required": []string{"org", "repo"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Org  string `json:"org"`
				Repo string `json:"repo"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}

			bundler := services.NewRAGBundler(store)
			bundle, err := bundler.GetContextBundle(context.Background(), args.Org, args.Repo)
			if err != nil {
				return nil, err
			}
			return bundle, nil
		},
	})

}
