// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package consumers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

// FetchSchema gets a schema from a URL, intercepting requests to the Substrate Public Registry.
func FetchSchema(ctx context.Context, url string, store db.Store) (string, error) {
	const registryPrefix = "registry.substrate.io/"

	if strings.HasPrefix(url, registryPrefix) {
		// Expect format: registry.substrate.io/@namespace/name/version
		path := strings.TrimPrefix(url, registryPrefix)
		parts := strings.Split(path, "/")
		if len(parts) != 3 || !strings.HasPrefix(parts[0], "@") {
			return "", fmt.Errorf("invalid registry url format: %s", url)
		}

		namespace := strings.TrimPrefix(parts[0], "@")
		name := parts[1]
		version := parts[2]

		schema, err := store.GetPublicSchema(ctx, namespace, name, version)
		if err != nil {
			return "", fmt.Errorf("failed to fetch schema from registry: %w", err)
		}

		return schema.SchemaContent, nil
	}

	// Fallback to standard HTTP fetch
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("invalid url: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch url %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status code %d when fetching %s", resp.StatusCode, url)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	return string(bodyBytes), nil
}
