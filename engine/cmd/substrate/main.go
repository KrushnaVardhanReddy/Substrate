package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/config"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/diff"
	initcmd "github.com/KrushnaVardhanReddy/substrate/engine/internal/init"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	sqlpkg "github.com/KrushnaVardhanReddy/substrate/engine/internal/sql"
	"github.com/KrushnaVardhanReddy/substrate/engine/pkg/ai"
	"github.com/spf13/cobra"
	"path/filepath"
)

var flattenAllOf bool
var configPath string
var format string
var schemaType string
var modeFlag string

func main() {
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
				base, err := sqlpkg.ParseSchema(basePath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
				head, err := sqlpkg.ParseSchema(revisionPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
				rep = sqlpkg.DiffSchemas(base, head)
			case "graphql":
				rep, err = diff.CompareGraphQL(basePath, revisionPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
			case "asyncapi":
				rep, err = diff.CompareAsyncAPI(basePath, revisionPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
			case "protobuf", "proto":
				rep, err = diff.CompareProto(basePath, revisionPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
			case "terraform-plan":
				rep, err = diff.CompareTerraformPlan(revisionPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
			case "ai-model":
				rep, err = diff.CompareAIML(basePath, revisionPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
			case "avro":
				rep, err = diff.CompareAvro(basePath, revisionPath, cfg)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					os.Exit(3)
				}
			default:
				rep, err = diff.CompareOpenAPI(basePath, revisionPath, flattenAllOf)
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
				if modeFlag == "strict" || modeFlag == "legacy" {
					finalMode = modeFlag
				} else {
					fmt.Fprintf(os.Stderr, "Error: invalid mode '%s'. Must be 'strict' or 'legacy'\n", modeFlag)
					os.Exit(3)
				}
			} else if cfg != nil && (cfg.Mode == "strict" || cfg.Mode == "legacy") {
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

				if len(rep.Warnings) > 0 {
					fmt.Printf("⚠️  WARNINGS (%d)\n", len(rep.Warnings))
					for _, chg := range rep.Warnings {
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
	diffCmd.Flags().StringVar(&modeFlag, "mode", "", "Execution mode: strict or legacy")

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

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
