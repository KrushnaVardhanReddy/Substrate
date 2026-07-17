package services

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
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
	consumerRepoID, err := store.UpsertRepo(ctx, orgID, req.ConsumerGithubRepoID, consumerName, req.ConsumerRepo, []byte("{}"))
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

		providerRepoID, err := store.UpsertRepo(ctx, orgID, dep.ProviderGithubRepoID, providerName, dep.ProviderRepo, []byte("{}"))
		if err != nil {
			return 0, err
		}

		// Fetch existing contracts to get the BASE schema before we overwrite it
		existingContracts, _ := store.GetContractsByProviderFullName(ctx, dep.ProviderRepo)
		var baseSchema string
		for _, c := range existingContracts {
			if c.SpecPath == dep.SpecPath {
				baseSchema = c.RawContent
				break
			}
		}

		contractID, err := store.UpsertContract(ctx, providerRepoID, dep.SchemaType, dep.SpecPath, dep.Branch, req.CommitSHA, dep.RawContent)
		if err != nil {
			return 0, err
		}

		// Perform Diff Analysis
		statusToSet := "SAFE"
		if baseSchema != "" {
			diffEngineURL := os.Getenv("DIFF_ENGINE_URL")
			if diffEngineURL == "" {
				diffEngineURL = "http://localhost:8080"
			}
			
			diffReq := DiffEngineRequest{
				BaseSchema: baseSchema,
				HeadSchema: dep.RawContent,
				SchemaType: dep.SchemaType,
			}
			
			if diffReqBytes, err := json.Marshal(diffReq); err == nil {
				log.Printf("ProcessSync: Calling DiffEngine at %s/diff", diffEngineURL)
				diffResp, err := http.Post(diffEngineURL+"/diff", "application/json", bytes.NewBuffer(diffReqBytes))
				if err != nil {
					log.Printf("ProcessSync: DiffEngine failed with error: %v", err)
				} else if diffResp.StatusCode != http.StatusOK {
					log.Printf("ProcessSync: DiffEngine returned status: %d", diffResp.StatusCode)
					diffResp.Body.Close()
				} else {
					var diffReport DiffReport
					if err := json.NewDecoder(diffResp.Body).Decode(&diffReport); err == nil {
						log.Printf("ProcessSync: DiffReport summary: BreakingCount=%d", diffReport.Summary.BreakingCount)
						if diffReport.Summary.BreakingCount > 0 {
							statusToSet = "BREAKING"
						}
					} else {
						log.Printf("ProcessSync: Failed to decode DiffReport: %v", err)
					}
					diffResp.Body.Close()
				}
			} else {
				log.Printf("ProcessSync: Failed to marshal diff request: %v", err)
			}
		}

		// For manual yaml configs, confidence score is implicitly 100 since it is explicitly declared
		if err := store.UpsertDependency(ctx, consumerRepoID, contractID, 100); err != nil {
			return 0, err
		}

		// Set the dependency status in the database based on the diff result
		_ = store.UpdateDependencyStatus(ctx, consumerRepoID, contractID, statusToSet)

		syncedCount++
	}

	return syncedCount, nil
}
