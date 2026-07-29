package services

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/config"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"gopkg.in/yaml.v3"
)

type DependencyPayload struct {
	ProviderRepo         string `json:"provider_repo"`
	ProviderGithubRepoID int64  `json:"provider_github_repo_id"`
	SchemaType           string `json:"schema_type"`
	SpecPath             string `json:"spec_path"`
	Branch               string `json:"branch"`
	RawContent           string `json:"raw_content"`
	RequiredNoticeDays   int    `json:"required_notice_days,omitempty"`
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

		// Try to extract metadata from the schema yaml block
		var parsed struct {
			Metadata *config.Metadata `yaml:"metadata"`
		}
		var providerMetadata json.RawMessage = []byte("{}")
		if err := yaml.Unmarshal([]byte(dep.RawContent), &parsed); err == nil && parsed.Metadata != nil {
			if metaBytes, err := json.Marshal(parsed.Metadata); err == nil {
				providerMetadata = json.RawMessage(metaBytes)
			}
		}

		providerRepoID, err := store.UpsertRepo(ctx, orgID, dep.ProviderGithubRepoID, providerName, dep.ProviderRepo, providerMetadata)
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
		if err := store.UpsertDependency(ctx, consumerRepoID, contractID, 100, dep.RequiredNoticeDays); err != nil {
			return 0, err
		}

		// Set the dependency status in the database based on the diff result
		_ = store.UpdateDependencyStatus(ctx, consumerRepoID, contractID, statusToSet)

		// Start a goroutine to ingest markdown guides
		// Pass a new context so it doesn't cancel when the request ends
		go IngestMarkdownGuides(context.Background(), store, req.Org, providerName)

		syncedCount++
	}

	return syncedCount, nil
}

// IngestMarkdownGuides clones the repo and upserts markdown guides
func IngestMarkdownGuides(ctx context.Context, store db.Store, org, repo string) {
	// Create a temporary directory for cloning
	tmpDir, err := os.MkdirTemp("", "substrate-guides-*")
	if err != nil {
		log.Printf("IngestMarkdownGuides: failed to create tmp dir: %v", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	repoURL := "https://github.com/" + org + "/" + repo + ".git"

	// Use git clone --depth 1
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth", "1", repoURL, tmpDir)
	if err := cmd.Run(); err != nil {
		log.Printf("IngestMarkdownGuides: git clone failed for %s/%s: %v", org, repo, err)
		return
	}

	docsPath := filepath.Join(tmpDir, "docs")
	if _, err := os.Stat(docsPath); os.IsNotExist(err) {
		log.Printf("IngestMarkdownGuides: no docs/ directory found in %s/%s", org, repo)
		return
	}

	// Walk docs directory
	err = filepath.WalkDir(docsPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		// Max size 500KB
		if info.Size() > 500*1024 {
			log.Printf("IngestMarkdownGuides: skipping %s: exceeds 500KB limit", path)
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			log.Printf("IngestMarkdownGuides: failed to read file %s: %v", path, err)
			return nil
		}

		// Extract title
		title := extractTitle(string(content), d.Name())

		relPath, err := filepath.Rel(docsPath, path)
		if err != nil {
			relPath = d.Name()
		}

		// Upsert guide
		if err := store.UpsertRepoGuide(context.Background(), org, repo, relPath, title, string(content)); err != nil {
			log.Printf("IngestMarkdownGuides: failed to upsert guide %s: %v", relPath, err)
		}

		return nil
	})

	if err != nil {
		log.Printf("IngestMarkdownGuides: walk docs failed for %s/%s: %v", org, repo, err)
	}
}

func extractTitle(content, defaultTitle string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return defaultTitle
}
