package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"

	"github.com/KrushnaVardhanReddy/Substrate/api/internal/db"
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

// CrossRepoCheckHandler handles the /api/v1/cross-repo-check endpoint.
func CrossRepoCheckHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		expectedToken := os.Getenv("INTERNAL_SERVICE_TOKEN")
		if expectedToken == "" || authHeader != "Bearer "+expectedToken {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		var req CrossRepoCheckRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid json"})
			return
		}

		ctx := r.Context()

		diffEngineURL := os.Getenv("DIFF_ENGINE_URL")
		if diffEngineURL == "" {
			diffEngineURL = "http://localhost:8080"
		}

		contracts, err := store.GetContractsByProviderFullName(ctx, req.ProviderRepo)
		if err != nil {
			http.Error(w, `{"error": "internal error"}`, http.StatusInternalServerError)
			return
		}

		response := CrossRepoCheckResponse{
			IsSafe:  true,
			Results: []ConsumerResult{},
		}

		consumerSeen := make(map[string]bool)

		for _, contract := range contracts {
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
					// Extract the error message from the engine if possible
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}
