// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package postman

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Client struct {
	HTTPClient  *http.Client
	BaseURL     string
	APIKey      string
	WorkspaceID string
}

func NewClient() *Client {
	return &Client{
		HTTPClient:  http.DefaultClient,
		BaseURL:     "https://api.postman.com",
		APIKey:      os.Getenv("POSTMAN_API_KEY"),
		WorkspaceID: os.Getenv("POSTMAN_WORKSPACE_ID"),
	}
}

type importPayload struct {
	Type    string                 `json:"type"`
	Input   string                 `json:"input"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type collectionInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type getWorkspaceResponse struct {
	Workspace struct {
		Collections []collectionInfo `json:"collections"`
	} `json:"workspace"`
}

func (c *Client) SyncSchema(ctx context.Context, schema string) error {
	title := extractTitleFromSchema(schema)
	if title == "" {
		title = "Substrate Auto-Generated API"
	}
	targetName := "[Substrate] " + title

	// 1. If we have a WorkspaceID, get the list of collections to check for an existing one.
	if c.WorkspaceID != "" {
		collections, err := c.getCollectionsInWorkspace(ctx)
		if err != nil {
			return fmt.Errorf("failed to get workspace collections: %w", err)
		}

		for _, coll := range collections {
			if coll.Name == targetName {
				// 2. Delete the old collection
				if err := c.deleteCollection(ctx, coll.ID); err != nil {
					return fmt.Errorf("failed to delete old collection: %w", err)
				}
				break
			}
		}
	}

	// Adjust schema title to ensure it gets imported with the [Substrate] prefix.
	schema = overrideSchemaTitle(schema, targetName)

	// 3. Import the new collection
	payload := importPayload{
		Type:  "string",
		Input: schema,
		Options: map[string]interface{}{
			"folderStrategy": "Paths",
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := c.BaseURL + "/import/openapi"
	if c.WorkspaceID != "" {
		url += "?workspace=" + c.WorkspaceID
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create import request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		req.Header.Set("x-api-key", c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("import request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("import failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

func (c *Client) getCollectionsInWorkspace(ctx context.Context) ([]collectionInfo, error) {
	url := c.BaseURL + "/workspaces/" + c.WorkspaceID
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if c.APIKey != "" {
		req.Header.Set("x-api-key", c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var data getWorkspaceResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return data.Workspace.Collections, nil
}

func (c *Client) deleteCollection(ctx context.Context, id string) error {
	url := c.BaseURL + "/collections/" + id
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	if c.APIKey != "" {
		req.Header.Set("x-api-key", c.APIKey)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	return nil
}

// extractTitleFromSchema makes a best effort to parse the title from a JSON or YAML OpenAPI schema.
func extractTitleFromSchema(schema string) string {
	var data struct {
		Info struct {
			Title string `json:"title"`
		} `json:"info"`
	}

	// Try parsing as JSON
	if err := json.Unmarshal([]byte(schema), &data); err == nil && data.Info.Title != "" {
		return data.Info.Title
	}

	// Basic YAML extraction logic
	lines := strings.Split(schema, "\n")
	inInfo := false
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, "info:") {
			inInfo = true
			continue
		}
		if inInfo {
			if !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				inInfo = false // Info block ended
				continue
			}
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "title:") {
				parts := strings.SplitN(trimmed, ":", 2)
				if len(parts) == 2 {
					title := strings.TrimSpace(parts[1])
					title = strings.Trim(title, "\"'")
					return title
				}
			}
		}
	}

	return ""
}

// overrideSchemaTitle tries to replace the "title": "..." field in a JSON or YAML schema with the target name.
func overrideSchemaTitle(schema, targetName string) string {
	var raw map[string]interface{}

	// If it's valid JSON
	if err := json.Unmarshal([]byte(schema), &raw); err == nil {
		if info, ok := raw["info"].(map[string]interface{}); ok {
			info["title"] = targetName
			raw["info"] = info
		} else {
			raw["info"] = map[string]interface{}{"title": targetName}
		}
		if modified, err := json.Marshal(raw); err == nil {
			return string(modified)
		}
		return schema
	}

	// For YAML, do a basic string replacement
	lines := strings.Split(schema, "\n")
	inInfo := false
	replaced := false
	var out []string

	for _, line := range lines {
		lineTrimmed := strings.TrimRight(line, "\r")
		if strings.HasPrefix(lineTrimmed, "info:") {
			inInfo = true
			out = append(out, line)
			continue
		}
		if inInfo {
			if !strings.HasPrefix(lineTrimmed, " ") && !strings.HasPrefix(lineTrimmed, "\t") {
				inInfo = false
			} else {
				trimmed := strings.TrimSpace(lineTrimmed)
				if strings.HasPrefix(trimmed, "title:") && !replaced {
					indent := line[:strings.Index(line, "title:")]
					out = append(out, indent+"title: "+targetName)
					replaced = true
					continue
				}
			}
		}
		out = append(out, line)
	}

	if replaced {
		return strings.Join(out, "\n")
	}

	// If info block didn't exist or title wasn't found in YAML, just return original
	return schema
}
