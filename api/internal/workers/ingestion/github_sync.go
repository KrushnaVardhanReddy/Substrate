// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package ingestion

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

type PublicRepoConfig struct {
	Namespace  string
	Name       string
	Version    string
	SchemaType string
	URL        string
}

// GitHubSyncWorker syncs OpenAPI schemas from public GitHub repositories into the Substrate Public Registry.
type GitHubSyncWorker struct {
	Store      db.Store
	HTTPClient *http.Client
}

func (w *GitHubSyncWorker) SyncRepositories(ctx context.Context, repos []PublicRepoConfig) error {
	client := w.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	for _, repo := range repos {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, repo.URL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request for %s: %w", repo.URL, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("failed to fetch schema from %s: %w", repo.URL, err)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			resp.Body.Close()
			return fmt.Errorf("unexpected status %d from %s", resp.StatusCode, repo.URL)
		}

		bodyBytes, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return fmt.Errorf("failed to read response body for %s: %w", repo.URL, err)
		}

		err = w.Store.PublishPublicSchema(
			ctx,
			repo.Namespace,
			repo.Name,
			repo.Version,
			repo.SchemaType,
			string(bodyBytes),
		)
		if err != nil {
			return fmt.Errorf("failed to publish schema %s/%s/%s: %w", repo.Namespace, repo.Name, repo.Version, err)
		}
	}

	return nil
}
