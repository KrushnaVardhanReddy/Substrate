package services

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	pkgconsumers "github.com/KrushnaVardhanReddy/substrate/api/internal/consumers"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

type CrossRepoCheckRequest struct {
	InstallationID    int64  `json:"installation_id"`
	Org               string `json:"org"`
	ProviderRepo      string `json:"provider_repo"`
	HeadSchemaContent string `json:"head_schema_content"`
	SchemaType        string `json:"schema_type"`
	ConfigContent     string `json:"config_content,omitempty"`
}

type DiffEngineRequest struct {
	BaseSchema string `json:"base_schema"`
	HeadSchema string `json:"head_schema"`
	SchemaType string `json:"schema_type"`
	Config     string `json:"config,omitempty"`
}

type DiffReportSummary struct {
	BreakingCount int `json:"breaking_count"`
	WarningCount  int `json:"warning_count"`
	InfoCount     int `json:"info_count"`
}

type DiffReport struct {
	Breaking []interface{}     `json:"breaking_changes"`
	Warning  []interface{}     `json:"warnings"`
	Info     []interface{}     `json:"safe_changes"`
	Summary  DiffReportSummary `json:"summary"`
}

type ConsumerResult struct {
	ConsumerRepo string      `json:"consumer_repo"`
	Status       string      `json:"status"` // "breaking" or "safe"
	DiffReport   *DiffReport `json:"diff_report"`
}

type CrossRepoCheckResponse struct {
	TotalConsumers  int              `json:"total_consumers"`
	BrokenConsumers int              `json:"broken_consumers"`
	IsSafe          bool             `json:"is_safe"`
	Results         []ConsumerResult `json:"results"`
}

// PerformCrossRepoCheck extracts the core logic of CrossRepoCheckHandler.
func PerformCrossRepoCheck(ctx context.Context, store db.Store, req CrossRepoCheckRequest) (CrossRepoCheckResponse, error) {
	diffEngineURL := os.Getenv("DIFF_ENGINE_URL")
	if diffEngineURL == "" {
		diffEngineURL = "http://localhost:8080"
	}

	response := CrossRepoCheckResponse{
		IsSafe:  true,
		Results: []ConsumerResult{},
	}

	contracts, err := store.GetContractsByProviderFullName(ctx, req.ProviderRepo)
	if err != nil {
		return response, err
	}

	consumerSeen := make(map[string]bool)

	for _, contract := range contracts {
		if contract.SchemaType != req.SchemaType {
			continue
		}

		consumers, err := store.GetConsumersByProviderContract(ctx, contract.ID)
		if err != nil {
			continue
		}

		for _, consumer := range consumers {
			if consumerSeen[consumer.ConsumerFullName] {
				continue
			}
			consumerSeen[consumer.ConsumerFullName] = true
			response.TotalConsumers++

			diffReq := DiffEngineRequest{
				BaseSchema: consumer.ContractRawContent,
				HeadSchema: req.HeadSchemaContent,
				SchemaType: req.SchemaType,
				Config:     req.ConfigContent,
			}

			diffReqBytes, err := json.Marshal(diffReq)
			if err != nil {
				continue
			}

			diffResp, err := http.Post(diffEngineURL+"/diff", "application/json", bytes.NewBuffer(diffReqBytes))
			if err != nil {
				response.BrokenConsumers++
				response.IsSafe = false
				response.Results = append(response.Results, ConsumerResult{
					ConsumerRepo: consumer.ConsumerFullName,
					Status:       "error",
					DiffReport: &DiffReport{
						Breaking: []interface{}{map[string]interface{}{"severity": "BREAKING", "description": "Failed to connect to diff engine"}},
						Summary:  DiffReportSummary{BreakingCount: 1},
					},
				})
				continue
			}

			if diffResp.StatusCode != http.StatusOK {
				var errResp struct {
					Error string `json:"error"`
				}
				json.NewDecoder(diffResp.Body).Decode(&errResp)
				diffResp.Body.Close()

				errMsg := "Internal Engine Error"
				if errResp.Error != "" {
					errMsg = errResp.Error
				}

				response.BrokenConsumers++
				response.IsSafe = false
				response.Results = append(response.Results, ConsumerResult{
					ConsumerRepo: consumer.ConsumerFullName,
					Status:       "error",
					DiffReport: &DiffReport{
						Breaking: []interface{}{
							map[string]interface{}{
								"severity":    "BREAKING",
								"description": errMsg,
							},
						},
						Summary: DiffReportSummary{BreakingCount: 1},
					},
				})
				continue
			}

			var diffReport DiffReport
			if err := json.NewDecoder(diffResp.Body).Decode(&diffReport); err != nil {
				diffResp.Body.Close()
				response.BrokenConsumers++
				response.IsSafe = false
				response.Results = append(response.Results, ConsumerResult{
					ConsumerRepo: consumer.ConsumerFullName,
					Status:       "error",
					DiffReport: &DiffReport{
						Breaking: []interface{}{map[string]interface{}{"severity": "BREAKING", "description": "Malformed response from diff engine"}},
						Summary:  DiffReportSummary{BreakingCount: 1},
					},
				})
				continue
			}
			diffResp.Body.Close()

			// Consumer Manifest Pruning
			var finalBreaking []interface{}
			var hasManifest bool
			var manifest *pkgconsumers.Manifest

			manifestBytes, err := store.GetConsumerManifests(ctx, req.ProviderRepo, consumer.ConsumerFullName)
			if err == nil && len(manifestBytes) > 0 {
				m, parseErr := pkgconsumers.Parse(manifestBytes)
				if parseErr == nil {
					hasManifest = true
					manifest = m
				}
			}

			for _, b := range diffReport.Breaking {
				bcMap, ok := b.(map[string]interface{})
				if !ok {
					finalBreaking = append(finalBreaking, b)
					continue
				}

				ruleID, _ := bcMap["rule_id"].(string)
				if ruleID == "FIELD_REMOVED" && hasManifest {
					path, _ := bcMap["path"].(string)

					parts := strings.Split(path, ".")
					fieldName := parts[len(parts)-1]

					isConsumed := false
					for _, consumes := range manifest.Consumes {
						for _, f := range consumes.Fields {
							if f == fieldName {
								isConsumed = true
								break
							}
						}
						if isConsumed {
							break
						}
					}

					if !isConsumed {
						bcMap["severity"] = "SAFE_NO_CONSUMERS"
						if diffReport.Info == nil {
							diffReport.Info = []interface{}{}
						}
						diffReport.Info = append(diffReport.Info, bcMap)
						diffReport.Summary.BreakingCount--
						continue
					}
				}

				finalBreaking = append(finalBreaking, b)
			}

			diffReport.Breaking = finalBreaking

			status := "safe"
			if diffReport.Summary.BreakingCount > 0 {
				status = "breaking"
				response.BrokenConsumers++
				response.IsSafe = false
			}

			if diffReport.Breaking == nil {
				diffReport.Breaking = []interface{}{}
			}
			if diffReport.Warning == nil {
				diffReport.Warning = []interface{}{}
			}
			if diffReport.Info == nil {
				diffReport.Info = []interface{}{}
			}

			response.Results = append(response.Results, ConsumerResult{
				ConsumerRepo: consumer.ConsumerFullName,
				Status:       status,
				DiffReport:   &diffReport,
			})
		}
	}

	return response, nil
}
