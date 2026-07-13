package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/config"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/discovery"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/handlers"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/integrations/postman"
)

type PushPayload struct {
	InstallationID int64  `json:"installation_id"`
	Org            string `json:"org"`
	Repo           string `json:"repo"`
	GithubRepoID   int64  `json:"github_repo_id"`
	CommitSHA      string `json:"commit_sha"`
	Files          []File `json:"files"`
}

type File struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

// PushHandler handles the GitHub push webhook payload forwarded from the GitHub App
func PushHandler(store db.Store, ghClient github.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PushPayload
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid json"}`, http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		orgID, err := store.UpsertOrg(ctx, req.InstallationID, req.Org)
		if err != nil {
			http.Error(w, `{"error": "internal error on UpsertOrg"}`, http.StatusInternalServerError)
			return
		}

		consumerName := req.Repo
		parts := strings.Split(req.Repo, "/")
		if len(parts) == 2 {
			consumerName = parts[1]
		}

		consumerRepoID, err := store.UpsertRepo(ctx, orgID, req.GithubRepoID, consumerName, req.Repo)
		if err != nil {
			http.Error(w, `{"error": "internal error on UpsertRepo"}`, http.StatusInternalServerError)
			return
		}

		discoveredCount := 0

		var matchPatterns []string
		for _, file := range req.Files {
			if file.Path == "substrate.yaml" {
				cfg, err := config.Parse([]byte(file.Content))
				if err == nil && cfg.Discovery != nil && len(cfg.Discovery.MatchPatterns) > 0 {
					matchPatterns = cfg.Discovery.MatchPatterns
				}
				break
			}
		}

		envScanner := discovery.NewEnvScanner(matchPatterns)

		for _, file := range req.Files {
			var deps []discovery.DiscoveredDependency

			if strings.HasSuffix(file.Path, ".env.example") || strings.HasSuffix(file.Path, ".env.template") || strings.HasSuffix(file.Path, ".env.sample") {
				deps = append(deps, envScanner.ScanEnvFile(file.Content)...)
			} else if strings.Contains(file.Path, "docker-compose") {
				deps = append(deps, envScanner.ScanDockerCompose(file.Content)...)
			} else if strings.HasSuffix(file.Path, ".yaml") || strings.HasSuffix(file.Path, ".yml") {
				deps = append(deps, envScanner.ScanKubernetesManifest(file.Content)...)
			}

			for _, dep := range deps {
				providerRepoName := extractProviderFromURL(dep.VarValue)
				if providerRepoName == "" {
					continue
				}

				fullName := req.Org + "/" + providerRepoName
				pseudoID := int64(crc32.ChecksumIEEE([]byte(fullName)))
				// Make it negative to avoid colliding with real github IDs
				if pseudoID > 0 {
					pseudoID = -pseudoID
				}
				
				providerRepoID, err := store.UpsertRepo(ctx, orgID, pseudoID, providerRepoName, fullName)
				if err != nil {
					continue
				}

				contractID, err := store.UpsertContract(ctx, providerRepoID, "unknown", "discovered", req.CommitSHA, req.CommitSHA, dep.VarValue)
				if err != nil {
					continue
				}

				err = store.UpsertDependency(ctx, consumerRepoID, contractID, dep.ConfidenceScore)
				if err == nil {
					discoveredCount++
				}
			}
		}

		// Cross-Repo Autofix Logic
		providerContracts, _ := store.GetContractsByProviderFullName(ctx, req.Repo)
		diffEngineURL := os.Getenv("DIFF_ENGINE_URL")
		if diffEngineURL == "" {
			diffEngineURL = "http://localhost:8080"
		}

		for _, contract := range providerContracts {
			for _, file := range req.Files {
				if file.Path == contract.SpecPath {
					diffReq := handlers.DiffEngineRequest{
						BaseSchema: contract.RawContent,
						HeadSchema: file.Content,
						SchemaType: contract.SchemaType,
					}
					diffReqBytes, err := json.Marshal(diffReq)
					if err != nil {
						continue
					}

					diffResp, err := http.Post(diffEngineURL+"/diff", "application/json", bytes.NewBuffer(diffReqBytes))
					if err != nil || diffResp.StatusCode != http.StatusOK {
						if diffResp != nil {
							diffResp.Body.Close()
						}
						continue
					}

					var diffReport handlers.DiffReport
					if err := json.NewDecoder(diffResp.Body).Decode(&diffReport); err != nil {
						diffResp.Body.Close()
						continue
					}
					diffResp.Body.Close()

					if diffReport.Summary.BreakingCount > 0 {
						var breakingChanges []handlers.BreakingChange
						for _, b := range diffReport.Breaking {
							if bcMap, ok := b.(map[string]interface{}); ok {
								bc := handlers.BreakingChange{
									RuleID:      fmt.Sprint(bcMap["rule_id"]),
									Path:        fmt.Sprint(bcMap["path"]),
									Description: fmt.Sprint(bcMap["description"]),
									Severity:    fmt.Sprint(bcMap["severity"]),
								}
								breakingChanges = append(breakingChanges, bc)
							}
						}

						consumers, _ := store.GetConsumersByProviderContract(ctx, contract.ID)
						for _, consumer := range consumers {
							go func(c db.ConsumerDependency, changes []handlers.BreakingChange, currentSchema, proposedSchema, providerRepo, schemaType string) {
								bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
								defer cancel()

								parts := strings.Split(c.ConsumerFullName, "/")
								if len(parts) != 2 {
									return
								}
								owner, repo := parts[0], parts[1]

								consumerFilePath, err := ghClient.SearchCode(bgCtx, owner, repo, providerRepo)
								if err != nil {
									log.Printf("Failed to find file referencing %s in %s: %v", providerRepo, c.ConsumerFullName, err)
									return
								}

								consumerSourceCode, err := ghClient.GetFileContent(bgCtx, owner, repo, consumerFilePath)
								if err != nil {
									log.Printf("Failed to get file content: %v", err)
									return
								}

								autofixReq := handlers.AIAutofixRequest{
									ProviderRepo:       providerRepo,
									SchemaType:         schemaType,
									CurrentSchema:      currentSchema,
									ProposedSchema:     proposedSchema,
									BreakingChanges:    changes,
									ConsumerSourceCode: consumerSourceCode,
								}

								autofixResp, err := handlers.GenerateAutofixPatch(autofixReq)
								if err != nil {
									log.Printf("Failed to generate autofix patch: %v", err)
									return
								}

								title := fmt.Sprintf("chore(substrate): Auto-fix breaking change from upstream [%s]", providerRepo)
								prBody := fmt.Sprintf("Substrate AI detected a breaking change in %s and generated this patch to fix it.\n\n**Reasoning:**\n%s", providerRepo, autofixResp.Explanation)

								_, err = ghClient.CreateDraftPR(bgCtx, owner, repo, "substrate-autofix-"+fmt.Sprint(time.Now().Unix()), autofixResp.SafePatch, title, prBody)
								if err != nil {
									log.Printf("Failed to create draft PR: %v", err)
								}
							}(consumer, breakingChanges, contract.RawContent, file.Content, req.Repo, contract.SchemaType)
						}
					}
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]int{"discovered": discoveredCount})
	}
}

func extractProviderFromURL(urlStr string) string {
	urlStr = strings.TrimPrefix(urlStr, "http://")
	urlStr = strings.TrimPrefix(urlStr, "https://")

	parts := strings.Split(urlStr, ":")
	host := parts[0]

	hostParts := strings.Split(host, ".")
	if len(hostParts) > 0 {
		return hostParts[0]
	}
	return ""
}

type WebhookPayload struct {
	Event   string `json:"event"`
	Status  string `json:"status"` // "SAFE", "APPROVED", etc.
	Schema  string `json:"schema"` // The OpenAPI schema JSON/YAML string
	Type    string `json:"type"`   // "openapi", etc.
	Version string `json:"version"`
}

type SyncClient interface {
	SyncSchema(ctx context.Context, schema string) error
}

func Handler() http.HandlerFunc {
	return HandlerWithClient(postman.NewClient())
}

func HandlerWithClient(client SyncClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload WebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		if payload.Type == "openapi" && (payload.Status == "SAFE" || payload.Status == "APPROVED") {
			ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()
			err := client.SyncSchema(ctx, payload.Schema)
			if err != nil {
				http.Error(w, "failed to sync to postman", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
