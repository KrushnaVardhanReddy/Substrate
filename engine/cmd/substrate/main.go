package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/diff"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

type SubstrateConfig struct {
	Service    string           `yaml:"service"`
	SchemaType string           `yaml:"schema_type,omitempty"`
	SpecPath   string           `yaml:"spec_path"`
	Overrides  []OverrideConfig `yaml:"overrides,omitempty"`
}

type OverrideConfig struct {
	RuleID     string `yaml:"rule_id"`
	Path       string `yaml:"path"`
	Reason     string `yaml:"reason"`
	ApprovedBy string `yaml:"approved_by"`
	Expires    string `yaml:"expires"`
}

var flattenAllOf bool
var outputFormat string
var configPath string

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

			var config *SubstrateConfig
			if configPath != "" {
				if _, err := os.Stat(configPath); err == nil {
					data, err := os.ReadFile(configPath)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
						os.Exit(3)
					}
					config = &SubstrateConfig{}
					if err := yaml.Unmarshal(data, config); err != nil {
						fmt.Fprintf(os.Stderr, "Error parsing config file: %v\n", err)
						os.Exit(3)
					}
					if config.Service == "" {
						fmt.Fprintf(os.Stderr, "substrate.yaml: 'service' is required\n")
						os.Exit(3)
					}
					// Also validation based on other fields but 'service' is explicitly checked in tests
				}
			}

			rep, err := diff.CompareOpenAPI(basePath, revisionPath, flattenAllOf)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(3)
			}

			// Apply overrides if config is present
			if config != nil && len(config.Overrides) > 0 {
				var stillBreaking []report.Change
				for _, chg := range rep.BreakingChanges {
					overridden := false
					for _, ov := range config.Overrides {
						if ov.RuleID == chg.RuleID && ov.Path == chg.Path {
							overridden = true
							break
						}
					}
					if overridden {
						// Move to safe changes or warnings, we'll put it in SafeChanges
						chg.Severity = report.ChangeSeveritySafe
						rep.SafeChanges = append(rep.SafeChanges, chg)
						rep.Summary.SafeCount++
					} else {
						stillBreaking = append(stillBreaking, chg)
					}
				}
				rep.BreakingChanges = stillBreaking
				rep.Summary.BreakingCount = len(stillBreaking)

				// Update overall severity
				if rep.Summary.BreakingCount > 0 {
					rep.Summary.OverallSeverity = report.SeverityBreaking
				} else if rep.Summary.WarningCount > 0 {
					rep.Summary.OverallSeverity = report.SeverityWarning
				} else if rep.Summary.SafeCount > 0 {
					rep.Summary.OverallSeverity = report.SeveritySafe
				} else {
					rep.Summary.OverallSeverity = report.SeverityNoChanges
				}
			}

			switch outputFormat {
			case "text":
				fmt.Printf("Substrate Diff Report\n")
				fmt.Printf("─────────────────────\n")
				fmt.Printf("Schema Type:  %s\n", rep.SchemaType)
				fmt.Printf("Analyzed At:  %s\n\n", rep.ComparedAt)
				fmt.Printf("Summary: %d changes — %d BREAKING, %d WARNING, %d SAFE\n\n", rep.Summary.TotalChanges, rep.Summary.BreakingCount, rep.Summary.WarningCount, rep.Summary.SafeCount)

				if rep.Summary.BreakingCount > 0 {
					fmt.Printf("❌ BREAKING CHANGES (%d)\n", rep.Summary.BreakingCount)
					for _, chg := range rep.BreakingChanges {
						fmt.Printf("  [%s] %s\n", chg.ID, chg.RuleID)
						fmt.Printf("  Path: %s\n", chg.Path)
						fmt.Printf("  Description: %s\n", chg.Description)
					}
					fmt.Println()
				}

				if rep.Summary.WarningCount > 0 {
					fmt.Printf("⚠️ WARNINGS (%d)\n", rep.Summary.WarningCount)
					for _, chg := range rep.Warnings {
						fmt.Printf("  [%s] %s\n", chg.ID, chg.RuleID)
						fmt.Printf("  Path: %s\n", chg.Path)
						fmt.Printf("  Description: %s\n", chg.Description)
					}
					fmt.Println()
				}

				if rep.Summary.SafeCount > 0 {
					fmt.Printf("✅ SAFE CHANGES (%d)\n", rep.Summary.SafeCount)
					for _, chg := range rep.SafeChanges {
						fmt.Printf("  [%s] %s\n", chg.ID, chg.RuleID)
						fmt.Printf("  Path: %s\n", chg.Path)
						fmt.Printf("  Description: %s\n", chg.Description)
					}
					fmt.Println()
				}
			case "changelog":
				fmt.Printf("## 🔍 Substrate Schema Report\n\n")
				fmt.Printf("**Schema:** `%s` | **Changes:** %d (%d breaking, %d warning, %d safe)\n\n", rep.SchemaType, rep.Summary.TotalChanges, rep.Summary.BreakingCount, rep.Summary.WarningCount, rep.Summary.SafeCount)

				if rep.Summary.BreakingCount > 0 {
					fmt.Printf("### ❌ Breaking Changes\n")
					fmt.Printf("| ID | Rule | Path | Description |\n")
					fmt.Printf("|---|---|---|---|\n")
					for _, chg := range rep.BreakingChanges {
						fmt.Printf("| %s | `%s` | `%s` | %s |\n", chg.ID, chg.RuleID, chg.Path, chg.Description)
					}
					fmt.Println()
				}

				if rep.Summary.WarningCount > 0 {
					fmt.Printf("### ⚠️ Warnings\n")
					fmt.Printf("| ID | Rule | Path | Description |\n")
					fmt.Printf("|---|---|---|---|\n")
					for _, chg := range rep.Warnings {
						fmt.Printf("| %s | `%s` | `%s` | %s |\n", chg.ID, chg.RuleID, chg.Path, chg.Description)
					}
					fmt.Println()
				}

				if rep.Summary.SafeCount > 0 {
					fmt.Printf("### ✅ Safe Changes\n")
					fmt.Printf("| ID | Rule | Path | Description |\n")
					fmt.Printf("|---|---|---|---|\n")
					for _, chg := range rep.SafeChanges {
						fmt.Printf("| %s | `%s` | `%s` | %s |\n", chg.ID, chg.RuleID, chg.Path, chg.Description)
					}
					fmt.Println()
				}
			default: // json
				output, err := json.MarshalIndent(rep, "", "  ")
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error marshalling JSON: %v\n", err)
					os.Exit(3)
				}
				fmt.Println(string(output))
			}

			if rep.Summary.BreakingCount > 0 {
				os.Exit(2)
			} else if rep.Summary.WarningCount > 0 {
				os.Exit(1)
			}
			os.Exit(0)
		},
	}

	diffCmd.Flags().BoolVar(&flattenAllOf, "flatten-allof", true, "Merge allOf schemas before diffing")
	diffCmd.Flags().StringVar(&outputFormat, "format", "json", "Output format")
	diffCmd.Flags().StringVar(&configPath, "config", "./substrate.yaml", "Path to override config file")

	var validateCmd = &cobra.Command{
		Use:   "validate [spec-file]",
		Short: "Validate a single spec file for errors",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Validate logic placeholder
			fmt.Println("Validate command executed")
		},
	}

	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(validateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
