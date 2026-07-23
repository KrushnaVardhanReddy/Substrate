package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var smellSchemaPath string

var schemaSmellCmd = &cobra.Command{
	Use:   "schema-smell",
	Short: "Detect schema smells in API designs via backend API",
	RunE: func(cmd *cobra.Command, args []string) error {
		if smellSchemaPath == "" {
			return fmt.Errorf("--schema is required")
		}

		schemaContent, err := os.ReadFile(smellSchemaPath)
		if err != nil {
			return fmt.Errorf("failed to read schema: %w", err)
		}

		apiURL := viper.GetString("SUBSTRATE_API_URL")
		if apiURL == "" {
			apiURL = "http://localhost:8090"
		}

		req, err := http.NewRequest("POST", apiURL+"/api/v1/schema/smell", bytes.NewReader(schemaContent))
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/yaml")
		apiToken := viper.GetString("REGISTRY_API_TOKEN")
		if apiToken != "" {
			req.Header.Set("Authorization", "Bearer "+apiToken)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to call API: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
		}

		body, _ := io.ReadAll(resp.Body)
		fmt.Println(string(body))
		return nil
	},
}

func init() {
	schemaSmellCmd.Flags().StringVar(&smellSchemaPath, "schema", "", "Path to the OpenAPI schema")
	viper.BindPFlags(schemaSmellCmd.Flags())
}
