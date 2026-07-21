package main

import (
	"fmt"
	"github.com/spf13/viper"
	"os"

	"github.com/KrushnaVardhanReddy/substrate/engine/pkg/fuzzer"
	"github.com/spf13/cobra"
)

var generateTestsBaseURL string

var generateTestsCmd = &cobra.Command{
	Use:   "generate-tests [spec-file]",
	Short: "Generate Go test code (using httpexpect) for OpenAPI constraints (e.g., maxLength, minimum, required)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		specPath := args[0]
		output, err := fuzzer.GenerateTests(specPath, generateTestsBaseURL)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating tests: %v\n", err)
			os.Exit(3)
		}
		fmt.Println(output)
	},
}

func init() {
	generateTestsCmd.Flags().StringVar(&generateTestsBaseURL, "url", "", "Base URL for the generated API tests")
	viper.BindPFlags(generateTestsCmd.Flags())
}
