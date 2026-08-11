// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package workers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/services"
	"github.com/google/uuid"
	"github.com/riverqueue/river"
)

func TestDiscoveryWorker_Work(t *testing.T) {
	mockStore := &db.MockStore{
		UpsertOrgFunc: func(ctx context.Context, installationID int64, orgName string) (uuid.UUID, error) {
			return uuid.New(), nil
		},
		UpsertRepoFunc: func(ctx context.Context, orgID uuid.UUID, githubRepoID int64, name, fullName string, metadata json.RawMessage) (uuid.UUID, error) {
			return uuid.New(), nil
		},
		UpsertContractFunc: func(ctx context.Context, providerRepoID uuid.UUID, schemaType, format, commitSha, version, rawContent string) (uuid.UUID, error) {
			return uuid.New(), nil
		},
		UpsertDependencyFunc: func(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, confidenceScore int, requiredNoticeDays int) error {
			return nil
		},
	}

	mockClient := &github.MockClient{
		GetFileContentFunc: func(ctx context.Context, owner, repo, path string) (string, error) {
			if path == "package.json" {
				return `{"dependencies": {"@myorg/backend-api": "1.0.0"}}`, nil
			}
			return "", nil
		},
	}

	worker := &DiscoveryWorker{
		Store:    mockStore,
		GHClient: mockClient,
	}

	job := &river.Job[DiscoveryJob]{
		Args: DiscoveryJob{
			Payload: services.PushPayload{
				InstallationID: 1,
				Org:            "myorg",
				Repo:           "myorg/frontend",
				GithubRepoID:   123,
				CommitSHA:      "abcdef",
			},
		},
	}

	err := worker.Work(context.Background(), job)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
