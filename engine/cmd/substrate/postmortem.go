package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/postmortem"
	"github.com/spf13/cobra"
)

var incidentDateArg string

var postmortemCmd = &cobra.Command{
	Use:   "postmortem",
	Short: "Generate an AI-driven blameless post-mortem for an incident",
	Run: func(cmd *cobra.Command, args []string) {
		registryApiToken := os.Getenv("REGISTRY_API_TOKEN")
		if registryApiToken == "" {
			fmt.Fprintln(os.Stderr, "Error: REGISTRY_API_TOKEN environment variable is not set")
			os.Exit(3)
		}

		if incidentDateArg == "" {
			fmt.Fprintln(os.Stderr, "Error: --incident flag is required")
			os.Exit(3)
		}

		incidentDate, err := time.Parse(time.RFC3339, incidentDateArg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: Invalid incident date format. Expected RFC3339 (e.g., 2023-10-10T12:00:00Z): %v\n", err)
			os.Exit(3)
		}

		apiURL := os.Getenv("SUBSTRATE_API_URL")
		if apiURL == "" {
			apiURL = "http://localhost:8090"
		}

		ctx := context.Background()
		err = postmortem.GeneratePostMortem(ctx, apiURL, registryApiToken, incidentDate, os.Stdout)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating post-mortem: %v\n", err)
			os.Exit(3)
		}
	},
}

func init() {
	postmortemCmd.Flags().StringVar(&incidentDateArg, "incident", "", "The incident date in RFC3339 format (e.g., 2023-10-10T12:00:00Z)")
}
