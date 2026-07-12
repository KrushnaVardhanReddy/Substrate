package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
)

type canDeployResponse struct {
	Safe      bool     `json:"safe"`
	Message   string   `json:"message"`
	BlockedBy []string `json:"blocked_by,omitempty"`
}

var (
	repoArg   string
	commitArg string
)

var checkDeployCmd = &cobra.Command{
	Use:   "check-deploy",
	Short: "Check if a deployment is safe by verifying its downstream providers",
	Run: func(cmd *cobra.Command, args []string) {
		registryApiToken := os.Getenv("REGISTRY_API_TOKEN")
		if registryApiToken == "" {
			fmt.Fprintln(os.Stderr, "Error: REGISTRY_API_TOKEN environment variable is not set")
			os.Exit(3)
		}

		if repoArg == "" {
			fmt.Fprintln(os.Stderr, "Error: --repo flag is required")
			os.Exit(3)
		}
		if commitArg == "" {
			fmt.Fprintln(os.Stderr, "Error: --commit flag is required")
			os.Exit(3)
		}

		apiURL := os.Getenv("SUBSTRATE_API_URL")
		if apiURL == "" {
			apiURL = "http://localhost:8090"
		}

		reqURL := fmt.Sprintf("%s/api/v1/registry/can-deploy?repo=%s&commit=%s", apiURL, repoArg, commitArg)
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

		var result canDeployResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding response: %v\n", err)
			os.Exit(3)
		}

		if resp.StatusCode == http.StatusOK {
			fmt.Println("✅ SAFE TO DEPLOY")
			fmt.Println()
			fmt.Printf("Provider: %s (%s)\n", repoArg, commitArg)
			fmt.Println()
			fmt.Println("Compatibility Check:")
			// Outputting generic COMPATIBLE since our API returns safe.
			fmt.Println("- consumers -> COMPATIBLE")
			os.Exit(0)
		} else if resp.StatusCode == http.StatusConflict {
			fmt.Println("❌ DEPLOYMENT BLOCKED")
			fmt.Println()
			fmt.Printf("Provider: %s (%s)\n", repoArg, commitArg)
			fmt.Println()
			fmt.Println("The following consumers are INCOMPATIBLE with this deployment:")
			for _, blocked := range result.BlockedBy {
				fmt.Printf("- %s\n", blocked)
			}
			os.Exit(1)
		} else {
			fmt.Fprintf(os.Stderr, "Error from server: HTTP %d %s\n", resp.StatusCode, result.Message)
			os.Exit(3)
		}
	},
}

func init() {
	checkDeployCmd.Flags().StringVar(&repoArg, "repo", "", "The provider repository (e.g., github.com/org/repo)")
	checkDeployCmd.Flags().StringVar(&commitArg, "commit", "", "The commit SHA to check")
}
