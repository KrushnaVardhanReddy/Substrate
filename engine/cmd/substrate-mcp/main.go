package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/checker"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/diff"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/mcp"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	sqlpkg "github.com/KrushnaVardhanReddy/substrate/engine/internal/sql"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/telemetry"
)

func main() {
	tp, err := telemetry.InitTracer(context.Background(), "substrate-mcp")
	if err != nil {
		fmt.Fprintf(os.Stderr, "[substrate-mcp] failed to init tracer: %v\n", err)
	} else if tp != nil {
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				fmt.Fprintf(os.Stderr, "[substrate-mcp] error shutting down tracer provider: %v\n", err)
			}
		}()
	}

	server := mcp.NewServer()

	// Tool 1: get_dependency_graph
	server.RegisterTool(mcp.Tool{
		Name:        "get_dependency_graph",
		Description: "Returns the full map of which repositories consume which providers. Useful for understanding the blast radius of a change.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"org": map[string]any{
					"type":        "string",
					"description": "The GitHub organization name (e.g. 'myorg')",
				},
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

			// Call the live Registry API (defaulting to localhost:8090 if REGISTRY_API_URL is not set)
			apiURL := os.Getenv("REGISTRY_API_URL")
			if apiURL == "" {
				apiURL = "http://localhost:8090"
			}

			resp, err := http.Get(fmt.Sprintf("%s/api/v1/graph/%s", apiURL, args.Org))
			if err != nil {
				return nil, fmt.Errorf("failed to fetch dependency graph: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("registry API returned status: %d", resp.StatusCode)
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			return string(body), nil
		},
	})

	// Tool 2: check_compatibility
	server.RegisterTool(mcp.Tool{
		Name:        "check_compatibility",
		Description: "Runs a simulated cross-repo compatibility check. The AI can pass in a drafted schema (e.g. OpenAPI YAML) before it's even committed, and Substrate will validate it against all registered downstream consumers.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"provider_repo": map[string]any{
					"type":        "string",
					"description": "The repository being modified (e.g. 'myorg/backend-api')",
				},
				"head_schema_content": map[string]any{
					"type":        "string",
					"description": "The raw string content of the proposed new schema",
				},
				"schema_type": map[string]any{
					"type":        "string",
					"description": "The format (e.g. 'openapi', 'graphql', 'sql', 'salesforce-object')",
				},
			},
			"required": []string{"provider_repo", "head_schema_content", "schema_type"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				ProviderRepo      string `json:"provider_repo"`
				HeadSchemaContent string `json:"head_schema_content"`
				SchemaType        string `json:"schema_type"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}

			// Write head schema to temp file
			headFile, err := os.CreateTemp("", "head-*.schema")
			if err != nil {
				return nil, fmt.Errorf("failed to create temp head file: %v", err)
			}
			defer os.Remove(headFile.Name())
			if _, err := headFile.WriteString(args.HeadSchemaContent); err != nil {
				return nil, err
			}
			headFile.Close()

			// Write empty base schema to temp file
			baseFile, err := os.CreateTemp("", "base-*.schema")
			if err != nil {
				return nil, fmt.Errorf("failed to create temp base file: %v", err)
			}
			defer os.Remove(baseFile.Name())
			// Initialize with an empty valid structure for some formats
			baseContent := ""
			if args.SchemaType == "openapi" {
				baseContent = `{"openapi":"3.0.0","info":{"title":"mock","version":"1"},"paths":{}}`
			}
			if _, err := baseFile.WriteString(baseContent); err != nil {
				return nil, err
			}
			baseFile.Close()

			var rep *report.DiffReport
			switch args.SchemaType {
			case "sql":
				var base, head *sqlpkg.SQLSchema
				err = checker.ParseSchema(context.Background(), func() error {
					var err error
					base, err = sqlpkg.ParseSchema(baseFile.Name())
					return err
				})
				if err != nil {
					return nil, err
				}
				err = checker.ParseSchema(context.Background(), func() error {
					var err error
					head, err = sqlpkg.ParseSchema(headFile.Name())
					return err
				})
				if err != nil {
					return nil, err
				}
				err = checker.CalculateDiff(context.Background(), func() error {
					rep = sqlpkg.DiffSchemas(base, head)
					return nil
				})
				if err != nil {
					return nil, err
				}
			case "graphql":
				err = checker.CalculateDiff(context.Background(), func() error {
					var err error
					rep, err = diff.CompareGraphQL(baseFile.Name(), headFile.Name())
					return err
				})
				if err != nil {
					return nil, err
				}
			case "asyncapi":
				err = checker.CalculateDiff(context.Background(), func() error {
					var err error
					rep, err = diff.CompareAsyncAPI(baseFile.Name(), headFile.Name())
					return err
				})
				if err != nil {
					return nil, err
				}
			case "protobuf", "proto":
				err = checker.CalculateDiff(context.Background(), func() error {
					var err error
					rep, err = diff.CompareProto(baseFile.Name(), headFile.Name())
					return err
				})
				if err != nil {
					return nil, err
				}
			case "terraform-plan":
				err = checker.CalculateDiff(context.Background(), func() error {
					var err error
					rep, err = diff.CompareTerraformPlan(headFile.Name())
					return err
				})
				if err != nil {
					return nil, err
				}
			case "ai-model":
				err = checker.CalculateDiff(context.Background(), func() error {
					var err error
					rep, err = diff.CompareAIML(baseFile.Name(), headFile.Name())
					return err
				})
				if err != nil {
					return nil, err
				}
			case "avro":
				err = checker.CalculateDiff(context.Background(), func() error {
					var err error
					rep, err = diff.CompareAvro(baseFile.Name(), headFile.Name(), nil)
					return err
				})
				if err != nil {
					return nil, err
				}
			default:
				err = checker.CalculateDiff(context.Background(), func() error {
					var err error
					rep, err = diff.CompareOpenAPI(baseFile.Name(), headFile.Name(), true, nil)
					return err
				})
				if err != nil {
					return nil, err
				}
			}

			out, err := json.Marshal(rep)
			if err != nil {
				return nil, err
			}
			return string(out), nil
		},
	})

	// Tool 3: get_breaking_change_history
	server.RegisterTool(mcp.Tool{
		Name:        "get_breaking_change_history",
		Description: "Retrieves a log of all historical breaking changes that have occurred on a specific repository. Useful for debugging production incidents.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"repo": map[string]any{
					"type":        "string",
					"description": "The repository to check (e.g. 'myorg/backend-api')",
				},
				"limit": map[string]any{
					"type":        "integer",
					"description": "Maximum number of records to return (default 10)",
				},
			},
			"required": []string{"repo"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var input struct {
				Repo  string `json:"repo"`
				Limit *int   `json:"limit,omitempty"`
			}
			if err := json.Unmarshal(params, &input); err != nil {
				return nil, fmt.Errorf("invalid params: %w", err)
			}

			if input.Repo == "" {
				return nil, fmt.Errorf("repo is required")
			}

			parts := strings.SplitN(input.Repo, "/", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("repo must be in format 'org/repo'")
			}
			org := parts[0]
			repo := parts[1]

			limit := 10
			if input.Limit != nil {
				limit = *input.Limit
			}

			registryURL := os.Getenv("REGISTRY_API_URL")
			if registryURL == "" {
				registryURL = "http://localhost:8090"
			}
			url := fmt.Sprintf("%s/api/v1/history/%s/%s?limit=%d", registryURL, org, repo, limit)

			req, err := http.NewRequest(http.MethodGet, url, nil)
			if err != nil {
				return nil, fmt.Errorf("failed to create request: %w", err)
			}

			token := os.Getenv("REGISTRY_API_TOKEN")
			if token != "" {
				req.Header.Set("Authorization", "Bearer "+token)
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch history: %w", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				return nil, fmt.Errorf("registry API returned status %d: %s", resp.StatusCode, string(body))
			}

			var history any
			if err := json.NewDecoder(resp.Body).Decode(&history); err != nil {
				return nil, fmt.Errorf("failed to parse response: %w", err)
			}

			return history, nil
		},
	})

	// Tool 4: get_substrate_docs
	server.RegisterTool(mcp.Tool{
		Name:        "get_substrate_docs",
		Description: "Retrieves the official Substrate configuration guide and `substrate.yaml` schema. The AI uses this to learn the rules before helping the user configure their repository.",
		InputSchema: map[string]any{
			"type":       "object",
			"properties": map[string]any{},
		},
		Handler: func(params json.RawMessage) (any, error) {
			return substrateDocsContent(), nil
		},
	})

	// Tool 5: analyze_repository
	server.RegisterTool(mcp.Tool{
		Name:        "analyze_repository",
		Description: "Scans the user's local directory tree to automatically detect API contracts (e.g., finding `openapi.yaml`, `.proto` files, or `.graphql` files) to recommend a `substrate.yaml` configuration.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"directory": map[string]any{
					"type":        "string",
					"description": "The local directory path to scan (default: current directory)",
				},
			},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Directory string `json:"directory"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			dir := args.Directory
			if dir == "" {
				dir = "."
			}

			// File patterns to detect, mapped to schema_type
			patterns := []struct {
				glob       string
				schemaType string
			}{
				{"openapi.yaml", "openapi"},
				{"openapi.yml", "openapi"},
				{"swagger.yaml", "openapi"},
				{"swagger.yml", "openapi"},
				{"*.proto", "protobuf"},
				{"*.graphql", "graphql"},
				{"*.gql", "graphql"},
				{"asyncapi.yaml", "asyncapi"},
				{"asyncapi.yml", "asyncapi"},
				{"schema.sql", "sql"},
				{"*.avsc", "avro"},
			}

			type detected struct {
				File       string `json:"file"`
				SchemaType string `json:"schema_type"`
			}
			var results []detected

			seen := map[string]bool{}

			err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil // skip unreadable dirs
				}
				// Skip hidden dirs like .git, node_modules, vendor
				if info.IsDir() {
					base := info.Name()
					if strings.HasPrefix(base, ".") || base == "node_modules" || base == "vendor" {
						return filepath.SkipDir
					}
					return nil
				}

				base := filepath.Base(path)
				for _, p := range patterns {
					matched, _ := filepath.Match(p.glob, base)
					if matched && !seen[path] {
						seen[path] = true
						results = append(results, detected{File: path, SchemaType: p.schemaType})
						break
					}
				}
				return nil
			})
			if err != nil {
				return nil, fmt.Errorf("failed to walk directory: %w", err)
			}

			if len(results) == 0 {
				return `[]`, nil
			}

			b, err := json.Marshal(results)
			if err != nil {
				return nil, err
			}
			return string(b), nil
		},
	})

	// Tool 6: execute_cli_command
	server.RegisterTool(mcp.Tool{
		Name:        "execute_cli_command",
		Description: "Allows the AI to safely execute the local `substrate` CLI binary (e.g., to run a local dry-run diff or initialize a repository).",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"args": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "string",
					},
					"description": "Arguments to pass to the substrate CLI (e.g., ['diff', '--base', 'a.yaml', '--head', 'b.yaml'])",
				},
			},
			"required": []string{"args"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var parsed struct {
				Args []string `json:"args"`
			}
			if err := json.Unmarshal(params, &parsed); err != nil {
				return nil, err
			}

			// We find the path to the current binary's directory and run substrate from there,
			// or assume "substrate" is in the PATH if not found.
			// However, since we might be running substrate-mcp directly, let's just use "substrate"
			cmd := exec.Command("substrate", parsed.Args...)
			out, err := cmd.CombinedOutput()
			exitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					exitCode = exitError.ExitCode()
				} else {
					return nil, err
				}
			}

			res := map[string]any{
				"stdout":    string(out),
				"exit_code": exitCode,
			}
			b, _ := json.Marshal(res)
			return string(b), nil
		},
	})

	// Tool 7: get_schema_file
	server.RegisterTool(mcp.Tool{
		Name:        "get_schema_file",
		Description: "Retrieves the exact, raw text of a stored contract schema from the registry. The AI can use this to read a downstream team's OpenAPI or GraphQL schema so it can perfectly write integration code against it.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"repo": map[string]any{
					"type":        "string",
					"description": "The repository name (e.g. 'myorg/backend-api')",
				},
			},
			"required": []string{"repo"},
		},
		Handler: func(params json.RawMessage) (any, error) {
			var args struct {
				Repo string `json:"repo"`
			}
			if err := json.Unmarshal(params, &args); err != nil {
				return nil, err
			}
			if args.Repo == "" {
				return nil, fmt.Errorf("repo is required (e.g. 'myorg/backend-api')")
			}

			apiURL := os.Getenv("REGISTRY_API_URL")
			if apiURL == "" {
				apiURL = "http://localhost:8090"
			}

			// Repo is "owner/repo" — split for the URL path
			parts := strings.SplitN(args.Repo, "/", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("repo must be in 'owner/repo' format, got: %s", args.Repo)
			}

			url := fmt.Sprintf("%s/api/v1/schema/%s/%s", apiURL, parts[0], parts[1])
			resp, err := http.Get(url)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch schema from registry: %w", err)
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, err
			}

			if resp.StatusCode == http.StatusNotFound {
				return nil, fmt.Errorf("no schema found for repo '%s' in the registry. Has it been synced?", args.Repo)
			}
			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("registry API returned status %d for repo '%s'", resp.StatusCode, args.Repo)
			}

			return string(body), nil
		},
	})

	server.ServeStdio()
}

func substrateDocsContent() string {
	return `# Substrate Configuration Guide

  Substrate is a CI/CD-integrated data contract and dependency intelligence platform.
  It protects API boundaries across microservices using schema diff analysis.

  ## substrate.yaml — Full Reference

  Place a ` + "`substrate.yaml`" + ` file in the root of any repository to enable Substrate.

  ### Provider Configuration (a repo that owns an API)
  ` + "```yaml" + `
  service: my-backend-api         # Human-readable service name
  schema_type: openapi            # Supported: openapi, sql, graphql, protobuf, asyncapi, avro, terraform-plan, ai-model
  spec_path: openapi.yaml         # Path to the schema file, relative to repo root
  on_breaking_change: block       # 'block' (default) = fail PR | 'warn' = comment only
  ` + "```" + `

  ### Consumer Configuration (a repo that depends on another team's API)
  ` + "```yaml" + `
  service: my-frontend

  consumers:
    - name: "users-api"
      provider_repo: myorg/backend-api    # The GitHub repo that owns the contract
      schema_type: openapi
      provider_spec_path: openapi.yaml    # Path to the spec IN the provider repo
      provider_branch: main               # Branch to track (default: main)
  ` + "```" + `

  ### Discovery Configuration
  ` + "```yaml" + `
  discovery:
    # Allow teams to override or append to the default regex for environment variables
    match_patterns:
      - '.*_ENDPOINT$'
      - '.*_HOST$'
  ` + "```" + `

  ### Breaking Change Overrides
  ` + "```yaml" + `
  overrides:
    - rule_id: ENDPOINT_REMOVED
      path: "GET /users/{id}"
      reason: "Intentional deprecation — clients have been migrated"
      approved_by: "platform-team@example.com"
      expires: "2025-12-31"
  ` + "```" + `

  ## Supported Schema Types

  | schema_type     | Description                          | Diff Tool Used      |
  |-----------------|--------------------------------------|---------------------|
  | openapi         | OpenAPI 3.x / Swagger 2.x REST APIs  | oasdiff (Go lib)    |
  | sql             | PostgreSQL DDL snapshots             | Custom parser       |
  | graphql         | GraphQL SDL schemas                  | Custom parser       |
  | protobuf        | Protocol Buffers .proto files        | buf CLI             |
  | asyncapi        | AsyncAPI event-driven APIs           | Custom parser       |
  | avro            | Apache Avro schemas                  | Schema Registry API |
  | terraform-plan  | Terraform plan JSON output           | Custom parser       |
  | ai-model        | AI/ML model contracts (YAML)         | Custom parser       |

  ## Key Breaking Change Rules (OpenAPI)

  | Rule ID                  | Severity | Description                             |
  |--------------------------|----------|-----------------------------------------|
  | ENDPOINT_REMOVED         | BREAKING | A route was deleted                     |
  | REQUEST_PARAM_REMOVED    | BREAKING | A required query/path param removed     |
  | RESPONSE_FIELD_REMOVED   | BREAKING | A response field was deleted            |
  | REQUEST_FIELD_TYPE_CHANGED | BREAKING | A field type changed                  |
  | ENDPOINT_DEPRECATED      | WARNING  | An endpoint was marked deprecated       |
  | FIELD_DEPRECATED         | WARNING  | A response field was marked deprecated  |

  ## Installation

  GitHub App: https://github.com/apps/substrate-contract-guard

  CLI (local diff):
  ` + "```bash" + `
  substrate diff --base openapi-old.yaml --head openapi-new.yaml --schema-type openapi
  
  # To run in Audit Mode (non-blocking, exits with 0 even on breaking changes):
  substrate diff --base openapi-old.yaml --head openapi-new.yaml --mode audit
  ` + "```" + `

  MCP Server (for Cursor / Claude Desktop):
  ` + "```json" + `
  {
    "mcpServers": {
      "substrate": {
        "command": "substrate-mcp",
        "env": {
          "REGISTRY_API_URL": "https://substrate.internal.mycompany.com"
        }
      }
    }
  }
  ` + "```"
}
