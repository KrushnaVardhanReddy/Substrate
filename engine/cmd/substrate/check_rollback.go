// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package main

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type canRollbackResponse struct {
	CanRollback bool     `json:"can_rollback"`
	Message     string   `json:"message"`
	BlockedBy   []string `json:"blocked_by,omitempty"`
}

var (
	rollbackRepoArg   string
	rollbackTargetSha string
)

var checkRollbackCmd = &cobra.Command{
	Use:   "check-rollback",
	Short: "Check if a rollback is safe by verifying its downstream providers",
	Run: func(cmd *cobra.Command, args []string) {
		registryApiToken := viper.GetString("REGISTRY_API_TOKEN")
		if registryApiToken == "" {
			fmt.Fprintln(os.Stderr, "Error: REGISTRY_API_TOKEN environment variable is not set")
			os.Exit(3)
		}

		if rollbackRepoArg == "" {
			fmt.Fprintln(os.Stderr, "Error: --repo flag is required")
			os.Exit(3)
		}
		if rollbackTargetSha == "" {
			fmt.Fprintln(os.Stderr, "Error: --target-sha flag is required")
			os.Exit(3)
		}

		parts := strings.Split(rollbackRepoArg, "/")
		if len(parts) != 2 {
			fmt.Fprintln(os.Stderr, "Error: invalid repo format, expected owner/repo")
			os.Exit(3)
		}
		org := parts[0]
		repo := parts[1]

		apiURL := viper.GetString("SUBSTRATE_API_URL")
		if apiURL == "" {
			apiURL = "http://localhost:8090"
		}

		reqURL := fmt.Sprintf("%s/api/v1/registry/can-rollback?org=%s&repo=%s&target_sha=%s", apiURL, org, repo, rollbackTargetSha)
		req, err := http.NewRequest("GET", reqURL, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
			os.Exit(3)
		}

		req.Header.Set("Authorization", "Bearer "+registryApiToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error making request: %v\n", err)
			os.Exit(3)
		}
		defer resp.Body.Close()

		var result canRollbackResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
			os.Exit(3)
		}

		if resp.StatusCode == http.StatusOK && result.CanRollback {
			fmt.Println("✅ SAFE TO ROLLBACK")
			fmt.Println()
			fmt.Printf("Provider: %s (%s)\n", rollbackRepoArg, rollbackTargetSha)
			fmt.Println()
			fmt.Println("Compatibility Check:")
			fmt.Println("- consumers -> COMPATIBLE")
			os.Exit(0)
		} else if resp.StatusCode == http.StatusConflict || !result.CanRollback {
			fmt.Println("❌ ROLLBACK BLOCKED")
			fmt.Println()
			fmt.Printf("Provider: %s (%s)\n", rollbackRepoArg, rollbackTargetSha)
			fmt.Println()
			fmt.Println("The following consumers are INCOMPATIBLE with this rollback:")
			for _, blocked := range result.BlockedBy {
				fmt.Printf("- %s\n", blocked)
			}
			os.Exit(2)
		} else {
			fmt.Fprintf(os.Stderr, "Error from server: HTTP %d %s\n", resp.StatusCode, result.Message)
			os.Exit(3)
		}
	},
}

func init() {
	checkRollbackCmd.Flags().StringVar(&rollbackRepoArg, "repo", "", "The provider repository (e.g., org/repo)")
	checkRollbackCmd.Flags().StringVar(&rollbackTargetSha, "target-sha", "", "The target commit SHA to rollback to")
	viper.BindPFlags(checkRollbackCmd.Flags())
}
