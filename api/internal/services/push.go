package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"hash/fnv"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/config"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/discovery"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/sandbox"
	"github.com/KrushnaVardhanReddy/substrate/engine/pkg/schemaowners"
)

type File struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type PushPayload struct {
	InstallationID int64  `json:"installation_id"`
	Org            string `json:"org"`
	Repo           string `json:"repo"`
	GithubRepoID   int64  `json:"github_repo_id"`
	CommitSHA      string `json:"commit_sha"`
	Files          []File `json:"files"`
}

func ProcessPush(ctx context.Context, store db.Store, ghClient github.Client, req PushPayload) (int, error) {
	// Fetch old provider contracts BEFORE we overwrite them in the DB!
	// This is required so DiffEngine can compare the OLD schema vs the NEW schema.
	providerContracts, _ := store.GetContractsByProviderFullName(ctx, req.Repo)

	// In a real app, we would resolve the organization name to an internal Org ID
	orgID, err := store.UpsertOrg(ctx, req.InstallationID, req.Org)
	if err != nil {
		return 0, fmt.Errorf("internal error on UpsertOrg: %w", err)
	}

	consumerName := req.Repo
	parts := strings.Split(req.Repo, "/")
	if len(parts) == 2 {
		consumerName = parts[1]
	}

	discoveredCount := 0

	var matchPatterns []string
	var repoMetadata json.RawMessage = json.RawMessage("{}")
	var newContractID uuid.UUID
	var parsedCfg *config.SubstrateConfig
	for _, file := range req.Files {
		if file.Path == "substrate.yaml" {
			cfg, err := config.Parse([]byte(file.Content))
			if err == nil {
				parsedCfg = cfg
				if cfg.Discovery != nil && len(cfg.Discovery.MatchPatterns) > 0 {
					matchPatterns = cfg.Discovery.MatchPatterns
				}
				if cfg.Metadata != nil {
					metaBytes, err := json.Marshal(cfg.Metadata)
					if err == nil {
						repoMetadata = json.RawMessage(metaBytes)
					}
				}
			}
			break
		}
	}

	repoID, err := store.UpsertRepo(ctx, orgID, req.GithubRepoID, consumerName, req.Repo, repoMetadata)
	if err != nil {
		return 0, fmt.Errorf("internal error on UpsertRepo: %w", err)
	}

	// If this repo acts as a provider, upsert its contract
	for _, file := range req.Files {
		if file.Path == "substrate.yaml" {
			cfg, err := config.Parse([]byte(file.Content))
			if err == nil && cfg.SchemaType != "" && cfg.HeadSchema != "" {
				var headContent string
				for _, f := range req.Files {
					if f.Path == cfg.HeadSchema {
						headContent = f.Content
						break
					}
				}
				if headContent != "" {
					var err error
					newContractID, err = store.UpsertContract(ctx, repoID, cfg.SchemaType, cfg.HeadSchema, "main", req.CommitSHA, headContent)
					if err == nil {
						for _, consumer := range cfg.Consumers {
							consumerOrgID, err := store.UpsertOrg(ctx, req.InstallationID, req.Org)
							if err == nil {
								// Generate a unique fake github_repo_id for the consumer
								fullName := req.Org + "/" + consumer.Name
								h := fnv.New32a()
								h.Write([]byte(fullName))
								fakeRepoID := -int64(h.Sum32())

								// consumer.Name is the repo name (e.g. frontend, emailservice)
								cRepoID, err := store.UpsertRepo(ctx, consumerOrgID, fakeRepoID, consumer.Name, fullName, json.RawMessage("{}"))
								if err == nil {
									err = store.UpsertDependency(ctx, cRepoID, newContractID, 1.0, 30)
									if err != nil {
										log.Printf("ProcessPush: Failed to UpsertDependency for consumer %s: %v", consumer.Name, err)
									} else {
										log.Printf("ProcessPush: Successfully upserted dependency for consumer %s", consumer.Name)
									}
								} else {
									log.Printf("ProcessPush: Failed to UpsertRepo for consumer %s: %v", consumer.Name, err)
								}
							} else {
								log.Printf("ProcessPush: Failed to UpsertOrg for consumer %s: %v", consumer.Name, err)
							}
						}
					}
				}
			}
			break
		}
	}

	consumerRepoID := repoID

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

			providerRepoID, err := store.UpsertRepo(ctx, orgID, pseudoID, providerRepoName, fullName, json.RawMessage("{}"))
			if err != nil {
				continue
			}

			contractID, err := store.UpsertContract(ctx, providerRepoID, "unknown", "discovered", "unknown", "unknown", dep.VarValue)
			if err != nil {
				continue
			}

			// For autodiscovered dependencies, required notice days defaults to 0
			err = store.UpsertDependency(ctx, consumerRepoID, contractID, dep.ConfidenceScore, 0)
			if err == nil {
				discoveredCount++
			}
		}
	}

	// Cross-Repo Autofix Logic
	// We comment out the Autofix generation to remove the import cycle to handlers for now.
	// Since this is offloading logic to services, we would typically have GenerateAutofixPatch inside services
	// but it's part of handlers currently. Let's not call GenerateAutofixPatch directly here for now, or we can move it to services.
	diffEngineURL := os.Getenv("DIFF_ENGINE_URL")
	if diffEngineURL == "" {
		diffEngineURL = "http://localhost:8080"
	}

	for _, contract := range providerContracts {
		for _, file := range req.Files {
			if file.Path == contract.SpecPath {
				diffReq := DiffEngineRequest{
					BaseSchema: contract.RawContent,
					HeadSchema: file.Content,
					SchemaType: contract.SchemaType,
				}
				diffReqBytes, err := json.Marshal(diffReq)
				if err != nil {
					log.Printf("ProcessPush: Failed to marshal diff request: %v", err)
					continue
				}

				log.Printf("ProcessPush: Calling DiffEngine at %s/diff for %s", diffEngineURL, contract.SpecPath)
				diffResp, err := http.Post(diffEngineURL+"/diff", "application/json", bytes.NewBuffer(diffReqBytes))
				if err != nil || diffResp.StatusCode != http.StatusOK {
					log.Printf("ProcessPush: DiffEngine failed. err=%v, status=%v", err, diffResp)
					if diffResp != nil {
						diffResp.Body.Close()
					}
					continue
				}

				var diffReport DiffReport
				if err := json.NewDecoder(diffResp.Body).Decode(&diffReport); err != nil {
					log.Printf("ProcessPush: Failed to decode diff report: %v", err)
					diffResp.Body.Close()
					continue
				}
				diffResp.Body.Close()

				log.Printf("ProcessPush: DiffReport summary: BreakingCount=%d", diffReport.Summary.BreakingCount)

				statusToSet := "SAFE"
				if diffReport.Summary.BreakingCount > 0 {
					statusToSet = "BREAKING"

					// Set the initial GitHub Check Suite status to pending
					parts := strings.Split(req.Repo, "/")
					if len(parts) == 2 {
						_ = ghClient.CreatePendingCheckRun(
							ctx,
							parts[0],
							parts[1],
							req.CommitSHA,
							"Substrate Contract Negotiation",
							"Waiting for Consumer Approval",
							"A breaking change was detected. Waiting for affected consumers to approve.",
						)
					}
				}

				// Update the status on the NEW contractID that we just created
				consumers, err := store.GetConsumersByProviderContract(ctx, newContractID)
				log.Printf("ProcessPush: GetConsumersByProviderContract returned %d consumers for new contract %v, err: %v", len(consumers), newContractID, err)
				for _, consumer := range consumers {
					consumerStatus := statusToSet
					if consumerStatus == "BREAKING" && parsedCfg != nil {
						// Check if this consumer has an override
						parts := strings.Split(consumer.ConsumerFullName, "/")
						consumerName := consumer.ConsumerFullName
						if len(parts) == 2 {
							consumerName = parts[1]
						}

						log.Printf("ProcessPush DEBUG: Checking overrides for consumer: %s (fullName: %s)", consumerName, consumer.ConsumerFullName)
						for _, cfgConsumer := range parsedCfg.Consumers {
							if cfgConsumer.Name == consumerName {
								log.Printf("ProcessPush DEBUG: Found matching consumer in parsedCfg: %s with %d overrides", cfgConsumer.Name, len(cfgConsumer.Overrides))
								for _, o := range cfgConsumer.Overrides {
									if o.RuleID == "*" {
										consumerStatus = "WARNING"
										log.Printf("ProcessPush DEBUG: Override matched! Setting status to WARNING")
										break
									}
								}
							}
						}
					}
					log.Printf("ProcessPush DEBUG: Updating dependency for %s to %s", consumer.ConsumerFullName, consumerStatus)
					_ = store.UpdateDependencyStatus(ctx, consumer.ConsumerRepoID, newContractID, consumerStatus)
				}

				if diffReport.Summary.BreakingCount > 0 {
					var breakingChanges []BreakingChange
					for _, b := range diffReport.Breaking {
						if bcMap, ok := b.(map[string]interface{}); ok {
							bc := BreakingChange{
								RuleID:      fmt.Sprint(bcMap["rule_id"]),
								Path:        fmt.Sprint(bcMap["path"]),
								Description: fmt.Sprint(bcMap["description"]),
								Severity:    fmt.Sprint(bcMap["severity"]),
							}
							breakingChanges = append(breakingChanges, bc)
						}
					}

					for _, consumer := range consumers {
						// Inside a worker context, we probably should either do this synchronously
						// or enqueue another job. For now, doing it synchronously inside the worker is fine
						// since it's background processed, but creating a draft PR could take a few seconds.
						func(c db.ConsumerDependency, changes []BreakingChange, currentSchema, proposedSchema, providerRepo, schemaType string) {
							bgCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
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

							autofixReq := AIAutofixRequest{
								ProviderRepo:       providerRepo,
								SchemaType:         schemaType,
								CurrentSchema:      currentSchema,
								ProposedSchema:     proposedSchema,
								BreakingChanges:    changes,
								ConsumerSourceCode: consumerSourceCode,
							}

							autofixResp, err := GenerateAutofixPatch(autofixReq)
							if err != nil {
								log.Printf("Failed to generate autofix patch: %v", err)
								return
							}

							title := fmt.Sprintf("chore(substrate): Auto-fix breaking change from upstream [%s]", providerRepo)

							breakingChangesJSON, _ := json.Marshal(changes)
							chaosSvc := NewChaosTestingService()
							chaosCode, err := chaosSvc.GenerateChaosTest(currentSchema, proposedSchema, string(breakingChangesJSON))
							chaosOutput := ""
							if err == nil {
								_, stderr, _ := sandbox.RunCode(chaosCode)
								chaosOutput = stderr
							}

							schemaOwnersContent, err := ghClient.GetFileContent(bgCtx, owner, repo, "SCHEMAOWNERS")
							var mentions []string
							if err == nil {
								parsed := schemaowners.Parse(schemaOwnersContent)
								for _, owners := range parsed {
									mentions = append(mentions, owners...)
								}
							}

							uniqueMentions := make(map[string]bool)
							var finalMentions []string
							for _, m := range mentions {
								if !uniqueMentions[m] {
									uniqueMentions[m] = true
									finalMentions = append(finalMentions, m)
								}
							}

							var prBody string
							if len(finalMentions) > 0 {
								prBody = fmt.Sprintf("Substrate AI detected a breaking change in %s and generated this patch to fix it.\n\n**Reasoning:**\n%s\n\n%s", providerRepo, autofixResp.Explanation, github.GenerateNegotiationComment(chaosOutput, finalMentions))
							} else {
								prBody = fmt.Sprintf("Substrate AI detected a breaking change in %s and generated this patch to fix it.\n\n**Reasoning:**\n%s\n\n%s", providerRepo, autofixResp.Explanation, github.GeneratePRComment(chaosOutput))
							}

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

	return discoveredCount, nil
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
