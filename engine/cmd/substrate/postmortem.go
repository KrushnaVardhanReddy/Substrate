// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package main

import (
	"bytes"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
)

var incidentDateArg string

var postmortemCmd = &cobra.Command{
	Use:   "postmortem",
	Short: "Generate an AI-driven blameless post-mortem for an incident",
	Run: func(cmd *cobra.Command, args []string) {
		registryApiToken := viper.GetString("REGISTRY_API_TOKEN")
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

		apiURL := viper.GetString("SUBSTRATE_API_URL")
		if apiURL == "" {
			apiURL = "http://localhost:8090"
		}

		reqBody := fmt.Sprintf(`{"incident_date":"%s"}`, incidentDate.Format(time.RFC3339))
		req, err := http.NewRequest("POST", apiURL+"/api/v1/postmortem", bytes.NewBufferString(reqBody))
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
			os.Exit(3)
		}
		req.Header.Set("Content-Type", "application/json")
		if registryApiToken != "" {
			req.Header.Set("Authorization", "Bearer "+registryApiToken)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating post-mortem: %v\n", err)
			os.Exit(3)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			fmt.Fprintf(os.Stderr, "API returned status %d: %s\n", resp.StatusCode, string(body))
			os.Exit(3)
		}

		io.Copy(os.Stdout, resp.Body)
	},
}

func init() {
	postmortemCmd.Flags().StringVar(&incidentDateArg, "incident", "", "The incident date in RFC3339 format (e.g., 2023-10-10T12:00:00Z)")
	viper.BindPFlags(postmortemCmd.Flags())
}
