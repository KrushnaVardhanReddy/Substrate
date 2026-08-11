// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"os"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var PluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage Substrate rulepack plugins",
}

var publishCmd = &cobra.Command{
	Use:   "publish [plugin-file.json]",
	Short: "Publish a rulepack plugin to the marketplace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filePath := args[0]
		data, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read plugin file: %w", err)
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(data, &payload); err != nil {
			return fmt.Errorf("invalid plugin json: %w", err)
		}

		apiURL := viper.GetString("SUBSTRATE_API_URL")
		if apiURL == "" {
			apiURL = "http://localhost:8090"
		}

		// Prepare publish request
		reqBody := map[string]interface{}{
			"name":           payload["name"],
			"description":    payload["description"],
			"schema_content": payload["schema"],
		}
		if reqBody["name"] == nil || reqBody["schema_content"] == nil {
			return fmt.Errorf("plugin must have 'name' and 'schema'")
		}

		reqJSON, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", apiURL+"/api/v1/plugins/publish", bytes.NewReader(reqJSON))
		if err != nil {
			return fmt.Errorf("failed to create publish request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		apiToken := viper.GetString("REGISTRY_API_TOKEN")
		if apiToken != "" {
			req.Header.Set("Authorization", "Bearer "+apiToken)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to publish plugin: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to publish: %s - %s", resp.Status, string(body))
		}

		fmt.Println("Plugin published successfully.")
		return nil
	},
}

var installCmd = &cobra.Command{
	Use:   "install [plugin-name]",
	Short: "Install a rulepack plugin from the marketplace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		pluginName := args[0]

		apiURL := viper.GetString("SUBSTRATE_API_URL")
		if apiURL == "" {
			apiURL = "http://localhost:8090"
		}

		// Fetch plugins
		resp, err := http.Get(apiURL + "/api/marketplace/plugins")
		if err != nil {
			return fmt.Errorf("failed to fetch plugins: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to fetch plugins: status %s", resp.Status)
		}

		var plugins []struct {
			Name          string          `json:"name"`
			SchemaContent json.RawMessage `json:"schema_content"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&plugins); err != nil {
			return fmt.Errorf("failed to parse plugins: %w", err)
		}

		var selectedPlugin *json.RawMessage
		for _, p := range plugins {
			if p.Name == pluginName {
				selectedPlugin = &p.SchemaContent
				break
			}
		}

		if selectedPlugin == nil {
			return fmt.Errorf("plugin '%s' not found", pluginName)
		}

		// Update substrate.yaml locally
		cfgPath := "substrate.yaml"
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Assuming plugin's schema contains custom rules
		var pluginSchema struct {
			Rules []config.CustomRule `json:"rules"`
		}
		if err := json.Unmarshal(*selectedPlugin, &pluginSchema); err != nil {
			return fmt.Errorf("invalid plugin schema: %w", err)
		}

		cfg.CustomRules = append(cfg.CustomRules, pluginSchema.Rules...)

		// Save back
		yamlData, err := yaml.Marshal(cfg)
		if err != nil {
			return fmt.Errorf("failed to serialize config: %w", err)
		}

		if err := os.WriteFile(cfgPath, yamlData, 0644); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		fmt.Printf("Plugin '%s' installed successfully.\n", pluginName)
		return nil
	},
}

func init() {
	PluginCmd.AddCommand(publishCmd)
	PluginCmd.AddCommand(installCmd)
}
