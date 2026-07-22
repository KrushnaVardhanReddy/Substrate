package services

import (
	"context"
	"encoding/json"
	"hash/crc32"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/discovery"
)

func ProcessDiscoveredEdges(ctx context.Context, store db.Store, req PushPayload, edges []discovery.DependencyEdge) error {
	orgID, err := store.UpsertOrg(ctx, req.InstallationID, req.Org)
	if err != nil {
		return err
	}

	consumerRepoID, err := store.UpsertRepo(ctx, orgID, req.GithubRepoID, req.Repo, req.Repo, json.RawMessage("{}"))
	if err != nil {
		return err
	}

	for _, edge := range edges {
		providerRepoName := extractProviderFromURL(edge.TargetURL)
		if providerRepoName == "" {
			providerRepoName = edge.TargetRepo
		}
		if providerRepoName == "" {
			continue
		}

		fullName := req.Org + "/" + providerRepoName
		pseudoID := int64(crc32.ChecksumIEEE([]byte(fullName)))
		if pseudoID > 0 {
			pseudoID = -pseudoID
		}

		providerRepoID, err := store.UpsertRepo(ctx, orgID, pseudoID, providerRepoName, fullName, json.RawMessage("{}"))
		if err != nil {
			continue
		}

		contractID, err := store.UpsertContract(ctx, providerRepoID, "unknown", "discovered", "unknown", "unknown", edge.TargetURL)
		if err != nil {
			continue
		}

		// Autodiscovered dependencies do not have SLAs configured, so pass 0
		_ = store.UpsertDependency(ctx, consumerRepoID, contractID, edge.Confidence, 0)
	}

	return nil
}
