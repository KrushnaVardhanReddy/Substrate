package services

import (
	"context"
	"strings"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

type DependencyPayload struct {
	ProviderRepo         string `json:"provider_repo"`
	ProviderGithubRepoID int64  `json:"provider_github_repo_id"`
	SchemaType           string `json:"schema_type"`
	SpecPath             string `json:"spec_path"`
	Branch               string `json:"branch"`
	RawContent           string `json:"raw_content"`
}

type SyncRequest struct {
	InstallationID       int64               `json:"installation_id"`
	Org                  string              `json:"org"`
	ConsumerRepo         string              `json:"consumer_repo"`
	ConsumerGithubRepoID int64               `json:"consumer_github_repo_id"`
	CommitSHA            string              `json:"commit_sha"`
	Dependencies         []DependencyPayload `json:"dependencies"`
}

func ProcessSync(ctx context.Context, store db.Store, req SyncRequest) (int, error) {
	// 1. UpsertOrg
	orgID, err := store.UpsertOrg(ctx, req.InstallationID, req.Org)
	if err != nil {
		return 0, err
	}

	// 2. UpsertRepo for consumer
	parts := strings.Split(req.ConsumerRepo, "/")
	consumerName := req.ConsumerRepo
	if len(parts) == 2 {
		consumerName = parts[1]
	}
	consumerRepoID, err := store.UpsertRepo(ctx, orgID, req.ConsumerGithubRepoID, consumerName, req.ConsumerRepo)
	if err != nil {
		return 0, err
	}

	// 3. For each dependency
	syncedCount := 0
	for _, dep := range req.Dependencies {
		providerParts := strings.Split(dep.ProviderRepo, "/")
		providerName := dep.ProviderRepo
		if len(providerParts) == 2 {
			providerName = providerParts[1]
		}

		providerRepoID, err := store.UpsertRepo(ctx, orgID, dep.ProviderGithubRepoID, providerName, dep.ProviderRepo)
		if err != nil {
			return 0, err
		}

		contractID, err := store.UpsertContract(ctx, providerRepoID, dep.SchemaType, dep.SpecPath, dep.Branch, req.CommitSHA, dep.RawContent)
		if err != nil {
			return 0, err
		}

		// For manual yaml configs, confidence score is implicitly 100 since it is explicitly declared
		if err := store.UpsertDependency(ctx, consumerRepoID, contractID, 100); err != nil {
			return 0, err
		}
		syncedCount++
	}

	return syncedCount, nil
}
