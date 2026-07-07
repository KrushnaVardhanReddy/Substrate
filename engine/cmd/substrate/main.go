package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "substrate",
		Short: "Substrate Diff Engine",
		Long:  `Substrate v0.1.0 — powered by oasdiff`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Substrate v0.1.0 — powered by oasdiff")
		},
	}

	var diffCmd = &cobra.Command{
		Use:   "diff [base-spec] [revision-spec]",
		Short: "Compare two schema files and output a DiffReport",
		Run: func(cmd *cobra.Command, args []string) {
			// Placeholder for diff command logic
			fmt.Println("Diff command executed")
		},
	}

	var validateCmd = &cobra.Command{
		Use:   "validate [spec-file]",
		Short: "Validate a single spec file for errors",
		Run: func(cmd *cobra.Command, args []string) {
			// Placeholder for validate command logic
			fmt.Println("Validate command executed")
		},
	}

	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(validateCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
