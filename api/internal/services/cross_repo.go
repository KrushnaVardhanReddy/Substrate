package services

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	pkgconsumers "github.com/KrushnaVardhanReddy/substrate/api/internal/consumers"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/crypto"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
	"github.com/KrushnaVardhanReddy/substrate/engine/crm"
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

type SLABreach struct {
	Consumer     string `json:"consumer"`
	RequiredDays int    `json:"required_days"`
}

type CrossRepoCheckResponse struct {
	TotalConsumers    int              `json:"total_consumers"`
	BrokenConsumers   int              `json:"broken_consumers"`
	IsSafe            bool             `json:"is_safe"`
	Results           []ConsumerResult `json:"results"`
	SLABreaches       []SLABreach      `json:"sla_breaches,omitempty"`
	AffectedCustomers int              `json:"affected_customers,omitempty"`
	AffectedMRR       float64          `json:"affected_mrr,omitempty"`
}

// PerformCrossRepoCheck extracts the core logic of CrossRepoCheckHandler.
func PerformCrossRepoCheck(ctx context.Context, store ports.CrossRepoStore, req CrossRepoCheckRequest) (CrossRepoCheckResponse, error) {
	diffEngineURL := os.Getenv("DIFF_ENGINE_URL")
	if diffEngineURL == "" {
		diffEngineURL = "http://localhost:8080"
	}

	response := CrossRepoCheckResponse{
		IsSafe:      true,
		Results:     []ConsumerResult{},
		SLABreaches: []SLABreach{},
	}

	// P10-T19: CRM/Billing Blast Radius
	var affectedCustomers int
	var affectedMRR float64
	var stripeClient *crm.StripeClient
	var sfClient *crm.SalesforceClient

	stripeKeyEnc, sfURLEnc, sfTokenEnc, sfClientIDEnc, sfClientSecretEnc, sfUserEnc, sfPassEnc, err := store.GetCRMSecrets(ctx, req.Org)
	if err == nil {
		mockKMS, _ := crypto.NewMockKMSClient("")
		// Initialize Stripe if configured
		if stripeKeyEnc != "" {
			key, err := mockKMS.Decrypt([]byte(stripeKeyEnc), "")
			if err == nil {
				stripeClient = crm.NewStripeClient(string(key))
			}
		}

		// Initialize Salesforce if configured
		if sfURLEnc != "" && sfTokenEnc != "" && sfClientIDEnc != "" && sfClientSecretEnc != "" && sfUserEnc != "" && sfPassEnc != "" {
			url, _ := mockKMS.Decrypt([]byte(sfURLEnc), "")
			token, _ := mockKMS.Decrypt([]byte(sfTokenEnc), "")
			clientID, _ := mockKMS.Decrypt([]byte(sfClientIDEnc), "")
			clientSecret, _ := mockKMS.Decrypt([]byte(sfClientSecretEnc), "")
			user, _ := mockKMS.Decrypt([]byte(sfUserEnc), "")
			pass, _ := mockKMS.Decrypt([]byte(sfPassEnc), "")
			sfClient, _ = crm.NewSalesforceClient(string(url), string(clientID), string(clientSecret), string(user), string(pass), string(token))
		}
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

				// If CRM clients are configured and there's a breaking change, compute blast radius
				if stripeClient != nil {
					// We'd ideally have a way to map consumer to stripe customer ID.
					// For demonstration in P10-T19, we will use a derived customer ID from consumer name
					// e.g., consumer.ConsumerFullName -> "cus_" + hash
					mrr, err := stripeClient.GetCustomerMRR(ctx, "cus_demo_"+consumer.ConsumerFullName)
					if err == nil {
						affectedCustomers++
						affectedMRR += mrr
					}
				} else if sfClient != nil {
					rev, err := sfClient.GetCustomerRevenue(ctx, "acc_demo_"+consumer.ConsumerFullName)
					if err == nil {
						affectedCustomers++
						affectedMRR += rev
					}
				}

				// Evaluate SLA Breach
				if consumer.RequiredNoticeDays > 0 {
					response.SLABreaches = append(response.SLABreaches, SLABreach{
						Consumer:     consumer.ConsumerFullName,
						RequiredDays: consumer.RequiredNoticeDays,
					})
				}
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

	if response.BrokenConsumers > 0 && affectedCustomers > 0 {
		response.AffectedCustomers = affectedCustomers
		response.AffectedMRR = affectedMRR
	}
	return response, nil
}
