package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/services"
	"github.com/google/uuid"
)

func RegisterTools(server *Server, store db.Store) {
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
}
