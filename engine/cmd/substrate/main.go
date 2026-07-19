package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"bytes"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/cache"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/checker"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/config"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/diff"
	initcmd "github.com/KrushnaVardhanReddy/substrate/engine/internal/init"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	sqlpkg "github.com/KrushnaVardhanReddy/substrate/engine/internal/sql"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/telemetry"
	"github.com/KrushnaVardhanReddy/substrate/engine/pkg/ai"
	"github.com/spf13/cobra"
	"net/http"
	"time"
)

var flattenAllOf bool
var configPath string
var format string
var schemaType string
var modeFlag string

func main() {
	tp, err := telemetry.InitTracer(context.Background(), "substrate-engine")
	if err != nil {
		log.Printf("[substrate-engine] failed to init tracer: %v\n", err)
	} else if tp != nil {
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				log.Printf("[substrate-engine] error shutting down tracer provider: %v", err)
			}
		}()
	}

	cacheDir := filepath.Join(os.Getenv("HOME"), ".substrate")
	os.MkdirAll(cacheDir, 0755)
	c, err := cache.InitCache(filepath.Join(cacheDir, "cache.db"))
	if err == nil && os.Getenv("SUBSTRATE_DISABLE_CACHE_SYNC") != "1" {
		go func() {
			apiURL := os.Getenv("SUBSTRATE_API_URL")
			if apiURL == "" {
				apiURL = "http://localhost:8090"
			}
			apiToken := os.Getenv("REGISTRY_API_TOKEN")
			org := os.Getenv("SUBSTRATE_ORG")
			if org == "" {
				org = "default"
			}
			for {
				c.SyncFromRemote(context.Background(), apiURL, apiToken, org)
				time.Sleep(1 * time.Minute)
			}
		}()
	}

	var rootCmd = &cobra.Command{
		Use:   "substrate",
		Short: "Substrate Diff Engine",
		Long:  `Substrate v0.1.0 — powered by oasdiff`,
	}

	var diffCmd = &cobra.Command{
		Use:   "diff [base-spec] [revision-spec]",
		Short: "Compare two schema files and output a DiffReport",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			basePath := args[0]
			revisionPath := args[1]

			cfg, err := config.LoadConfig(configPath)
			if err != nil {
				if !errors.Is(err, os.ErrNotExist) || configPath != "./substrate.yaml" {
					fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
					os.Exit(3)
				}
				// Proceed with defaults if default config file is missing
			}

			finalSchemaType := schemaType
			if finalSchemaType == "" {
				if cfg != nil && cfg.SchemaType != "" {
					finalSchemaType = cfg.SchemaType
				} else {
					ext := filepath.Ext(basePath)
					if ext == ".sql" {
						finalSchemaType = "sql"
					} else {
						finalSchemaType = "openapi"
					}
				}
			}

			var rep *report.DiffReport

			switch finalSchemaType {
			case "sql":
				var base, head *sqlpkg.SQLSchema
				err = checker.ParseSchema(context.Background(), func() error {
					var err error
					base, err = sqlpkg.ParseSchema(basePath)
					return err
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
				err = checker.ParseSchema(context.Background(), func() error {
					var err error
					head, err = sqlpkg.ParseSchema(revisionPath)
					return err
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
				err = checker.CalculateDiff(context.Background(), func() error {
					rep = sqlpkg.DiffSchemas(base, head)
					return nil
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
			case "graphql":
				err = checker.CalculateDiff(context.Background(), func() error {
					var err error
					rep, err = diff.CompareGraphQL(basePath, revisionPath)
					return err
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
			case "asyncapi":
				err = checker.CalculateDiff(context.Background(), func() error {
					var err error
					rep, err = diff.CompareAsyncAPI(basePath, revisionPath)
					return err
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
			case "protobuf", "proto":
				err = checker.CalculateDiff(context.Background(), func() error {
					var err error
					rep, err = diff.CompareProto(basePath, revisionPath)
					return err
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
			case "terraform-plan":
				err = checker.CalculateDiff(context.Background(), func() error {
					var err error
					rep, err = diff.CompareTerraformPlan(revisionPath)
					return err
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
			case "ai-model":
				err = checker.CalculateDiff(context.Background(), func() error {
					var err error
					rep, err = diff.CompareAIML(basePath, revisionPath)
					return err
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
			case "avro":
				err = checker.CalculateDiff(context.Background(), func() error {
					var err error
					rep, err = diff.CompareAvro(basePath, revisionPath, cfg)
					return err
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
			default:
				var rules []config.CustomRule
				if cfg != nil {
					rules = cfg.CustomRules
				}
				err = checker.CalculateDiff(context.Background(), func() error {
					var err error
					rep, err = diff.CompareOpenAPI(basePath, revisionPath, flattenAllOf, rules)
					return err
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
			}

			if cfg != nil {
				rep = diff.ApplyConfigAndTraffic(rep, cfg, "", "")
			}

			finalMode := "strict"
			if modeFlag != "" {
				if modeFlag == "strict" || modeFlag == "legacy" || modeFlag == "audit" {
					finalMode = modeFlag
				} else {
					fmt.Fprintf(os.Stderr, "Error: invalid mode '%s'. Must be 'strict', 'legacy', or 'audit'\n", modeFlag)
					os.Exit(3)
				}
			} else if cfg != nil && (cfg.Mode == "strict" || cfg.Mode == "legacy" || cfg.Mode == "audit") {
				finalMode = cfg.Mode
			}
			rep.Mode = finalMode

			if format == "json" {
				output, err := json.MarshalIndent(rep, "", "  ")
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error marshalling JSON: %v\n", err)
					os.Exit(3)
				}
				fmt.Println(string(output))

				// Post to API if configured
				apiURL := os.Getenv("SUBSTRATE_API_URL")
				apiToken := os.Getenv("REGISTRY_API_TOKEN")
				if apiURL != "" && apiToken != "" {
					payload := map[string]interface{}{
						"diff_report":   rep,
						"is_audit_mode": finalMode == "audit",
					}
					payloadBytes, _ := json.Marshal(payload)

					req, _ := http.NewRequest("POST", apiURL+"/api/v1/diff", bytes.NewBuffer(payloadBytes))
					req.Header.Set("Content-Type", "application/json")
					req.Header.Set("Authorization", "Bearer "+apiToken)

					client := &http.Client{}
					resp, err := client.Do(req)
					if err == nil {
						resp.Body.Close()
					}
				}
			} else if format == "text" {
				if finalMode == "legacy" && len(rep.BreakingChanges) > 0 {
					fmt.Println("⚠️ LEGACY MODE: Breaking changes detected, but merge is not blocked.")
				}
				fmt.Println("Substrate Diff Report")
				fmt.Println("─────────────────────")
				fmt.Printf("Schema Type:  %s\n", rep.SchemaType)
				fmt.Printf("Analyzed At:  %s\n\n", rep.ComparedAt)

				fmt.Printf("Summary: %d changes — %d BREAKING, %d WARNING, %d SAFE\n\n",
					rep.Summary.TotalChanges, rep.Summary.BreakingCount, rep.Summary.WarningCount, rep.Summary.SafeCount)

				if len(rep.BreakingChanges) > 0 {
					fmt.Printf("❌ BREAKING CHANGES (%d)\n", len(rep.BreakingChanges))
					for _, chg := range rep.BreakingChanges {
						fmt.Printf("  [%s] %s\n", chg.ID, chg.RuleID)
						fmt.Printf("  Path: %s\n", chg.Path)
						fmt.Printf("  Description: %s\n", chg.Description)
						if chg.Recommendation != nil {
							fmt.Printf("  Recommendation: %s\n", *chg.Recommendation)
						}
						fmt.Println()
					}
				}

				var perfRisks []report.Change
				var otherWarnings []report.Change
				for _, w := range rep.Warnings {
					if w.RuleID == "FOREIGN_KEY_MISSING_INDEX" || w.RuleID == "COLUMN_ADDED_WITH_DEFAULT" {
						perfRisks = append(perfRisks, w)
					} else {
						otherWarnings = append(otherWarnings, w)
					}
				}

				if len(perfRisks) > 0 {
					fmt.Printf("🚀 PERFORMANCE RISKS (%d)\n", len(perfRisks))
					for _, chg := range perfRisks {
						fmt.Printf("  [%s] %s\n", chg.ID, chg.RuleID)
						fmt.Printf("  Path: %s\n", chg.Path)
						fmt.Printf("  Description: %s\n", chg.Description)
						if chg.Recommendation != nil {
							fmt.Printf("  Recommendation: %s\n", *chg.Recommendation)
						}
						fmt.Println()
					}
				}

				if len(otherWarnings) > 0 {
					fmt.Printf("⚠️  WARNINGS (%d)\n", len(otherWarnings))
					for _, chg := range otherWarnings {
						fmt.Printf("  [%s] %s\n", chg.ID, chg.RuleID)
						fmt.Printf("  Path: %s\n", chg.Path)
						fmt.Printf("  Description: %s\n", chg.Description)
						if chg.Recommendation != nil {
							fmt.Printf("  Recommendation: %s\n", *chg.Recommendation)
						}
						fmt.Println()
					}
				}

				if len(rep.SafeChanges) > 0 {
					fmt.Printf("✅ SAFE CHANGES (%d)\n", len(rep.SafeChanges))
					for _, chg := range rep.SafeChanges {
						fmt.Printf("  [%s] %s\n", chg.ID, chg.RuleID)
						fmt.Printf("  Path: %s\n", chg.Path)
						fmt.Printf("  Description: %s\n", chg.Description)
						fmt.Println()
					}
				}

			} else if format == "changelog" {
				fmt.Println("## 🔍 Substrate Schema Report")
				fmt.Println()
				fmt.Printf("**Schema:** `%s` | **Changes:** %d (%d breaking, %d warning, %d safe)\n\n",
					rep.SchemaType, rep.Summary.TotalChanges, rep.Summary.BreakingCount, rep.Summary.WarningCount, rep.Summary.SafeCount)

				if len(rep.BreakingChanges) > 0 {
					fmt.Println("### ❌ Breaking Changes")
					fmt.Println("| ID | Rule | Path | Description |")
					fmt.Println("|---|---|---|---|")
					for _, chg := range rep.BreakingChanges {
						fmt.Printf("| %s | `%s` | `%s` | %s |\n", chg.ID, chg.RuleID, chg.Path, chg.Description)
					}
					fmt.Println()
					for _, chg := range rep.BreakingChanges {
						if chg.Recommendation != nil {
							fmt.Printf("> **Recommendation:** %s\n\n", *chg.Recommendation)
						}
					}
				}

				if len(rep.Warnings) > 0 {
					fmt.Println("### ⚠️ Warnings")
					fmt.Println("| ID | Rule | Path | Description |")
					fmt.Println("|---|---|---|---|")
					for _, chg := range rep.Warnings {
						fmt.Printf("| %s | `%s` | `%s` | %s |\n", chg.ID, chg.RuleID, chg.Path, chg.Description)
					}
					fmt.Println()
				}

				if len(rep.SafeChanges) > 0 {
					fmt.Println("### ✅ Safe Changes")
					fmt.Println("| ID | Rule | Path | Description |")
					fmt.Println("|---|---|---|---|")
					for _, chg := range rep.SafeChanges {
						fmt.Printf("| %s | `%s` | `%s` | %s |\n", chg.ID, chg.RuleID, chg.Path, chg.Description)
					}
					fmt.Println()
				}

				fmt.Println("---\n*Generated by [Substrate](https://github.com/KrushnaVardhanReddy/substrate) — schema contract protection*")
			} else {
				fmt.Fprintf(os.Stderr, "Unknown format: %s\n", format)
				os.Exit(3)
			}

			if finalMode == "legacy" {
				os.Exit(0)
			} else if finalMode == "audit" {
				if rep.Summary.BreakingCount > 0 {
					fmt.Fprintln(os.Stderr, "[AUDIT MODE] Breaking changes detected, but exiting with 0 to allow merge.")
				}
				os.Exit(0)
			} else {
				if rep.Summary.BreakingCount > 0 {
					os.Exit(2)
				} else if rep.Summary.WarningCount > 0 {
					os.Exit(1)
				}
				os.Exit(0)
			}
		},
	}

	diffCmd.Flags().BoolVar(&flattenAllOf, "flatten-allof", true, "Merge allOf schemas before diffing")
	diffCmd.Flags().StringVar(&configPath, "config", "./substrate.yaml", "Path to override config file")
	diffCmd.Flags().StringVar(&format, "format", "json", "Output format")
	diffCmd.Flags().StringVar(&schemaType, "schema-type", "", "Force schema type")
	diffCmd.Flags().StringVar(&modeFlag, "mode", "", "Execution mode: strict, legacy, or audit")

	var validateCmd = &cobra.Command{
		Use:   "validate [spec-file]",
		Short: "Validate a single spec file for errors",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Validate logic placeholder
			fmt.Println("Validate command executed")
		},
	}

	var designFlag bool
	var initOptions initcmd.InitOptions
	var initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize Substrate in the current directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			if designFlag {
				err := ai.RunArchitect(os.Stdin, os.Stdout)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Architect Error: %v\n", err)
					os.Exit(3)
				}
				// The architect generates openapi.yaml
				initOptions.Spec = "openapi.yaml"
			}

			err := initcmd.Init(initOptions)
			if err != nil {
				if errors.Is(err, initcmd.ErrPermissionDenied) {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(1)
				}
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(3)
			}
			return nil
		},
	}

	initCmd.Flags().StringVar(&initOptions.Service, "service", "", "The service name written into substrate.yaml")
	initCmd.Flags().StringVar(&initOptions.Spec, "spec", "", "Path to the OpenAPI spec file, relative to repo root")
	initCmd.Flags().StringVar(&initOptions.SchemaType, "schema-type", "openapi", "Schema type: openapi, sql, graphql, protobuf")
	initCmd.Flags().StringVar(&initOptions.Branch, "branch", "main", "The protected branch to check PRs against")
	initCmd.Flags().BoolVar(&initOptions.Force, "force", false, "Overwrite existing files without prompting")
	initCmd.Flags().BoolVar(&initOptions.NoWorkflow, "no-workflow", false, "Skip generating .github/workflows/substrate.yml")
	initCmd.Flags().BoolVar(&initOptions.NoConfig, "no-config", false, "Skip generating substrate.yaml")
	initCmd.Flags().BoolVar(&designFlag, "design", false, "Start AI architect to scaffold your API spec")

	var port string
	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start HTTP mode (Container Service)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runServe(port); err != nil {
				log.Fatal(err)
			}
			return nil
		},
	}
	serveCmd.Flags().StringVar(&port, "port", "8080", "Port to listen on")

	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(generateTestsCmd)
	rootCmd.AddCommand(mockCmd)
	rootCmd.AddCommand(checkDeployCmd)
	rootCmd.AddCommand(checkRollbackCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
