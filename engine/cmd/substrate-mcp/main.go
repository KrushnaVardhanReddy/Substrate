package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/diff"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/mcp"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	sqlpkg "github.com/KrushnaVardhanReddy/substrate/engine/internal/sql"
)

func main() {
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
				base, err := sqlpkg.ParseSchema(baseFile.Name())
				if err != nil {
					return nil, err
				}
				head, err := sqlpkg.ParseSchema(headFile.Name())
				if err != nil {
					return nil, err
				}
				rep = sqlpkg.DiffSchemas(base, head)
			case "graphql":
				rep, err = diff.CompareGraphQL(baseFile.Name(), headFile.Name())
				if err != nil {
					return nil, err
				}
			case "asyncapi":
				rep, err = diff.CompareAsyncAPI(baseFile.Name(), headFile.Name())
				if err != nil {
					return nil, err
				}
			case "protobuf", "proto":
				rep, err = diff.CompareProto(baseFile.Name(), headFile.Name())
				if err != nil {
					return nil, err
				}
			case "terraform-plan":
				rep, err = diff.CompareTerraformPlan(headFile.Name())
				if err != nil {
					return nil, err
				}
			case "ai-model":
				rep, err = diff.CompareAIML(baseFile.Name(), headFile.Name())
				if err != nil {
					return nil, err
				}
			case "avro":
				rep, err = diff.CompareAvro(baseFile.Name(), headFile.Name(), nil)
				if err != nil {
					return nil, err
				}
			default:
				rep, err = diff.CompareOpenAPI(baseFile.Name(), headFile.Name(), true)
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
			return `[{"timestamp": "2024-01-01T00:00:00Z", "git_sha": "mocksha", "breaking_changes": []}]`, nil
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
			return `# Substrate Configuration Guide\n\n...mock docs...`, nil
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
			return `[{"file": "openapi.yaml", "type": "openapi"}]`, nil
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
			return `{"schema": "mock raw schema", "format": "openapi"}`, nil
		},
	})

	server.ServeStdio()
}
