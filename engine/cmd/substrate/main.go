package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/diff"
	"github.com/spf13/cobra"
)

var flattenAllOf bool

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

			report, err := diff.CompareOpenAPI(basePath, revisionPath, flattenAllOf)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(3)
			}

			output, err := json.MarshalIndent(report, "", "  ")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error marshalling JSON: %v\n", err)
				os.Exit(3)
			}

			fmt.Println(string(output))

			if report.Summary.BreakingCount > 0 {
				os.Exit(2)
			} else if report.Summary.WarningCount > 0 {
				os.Exit(1)
			}
			os.Exit(0)
		},
	}

	diffCmd.Flags().BoolVar(&flattenAllOf, "flatten-allof", true, "Merge allOf schemas before diffing")

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
