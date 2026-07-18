package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/finops"
)

type PredictCostRequest struct {
	BaseSchema     map[string]interface{} `json:"base_schema"`
	ProposedSchema map[string]interface{} `json:"proposed_schema"`
	EndpointPath   string                 `json:"endpoint_path"`
}

type PredictCostResponse struct {
	BaseBytes       int64   `json:"base_bytes"`
	ProposedBytes   int64   `json:"proposed_bytes"`
	MonthlyCostDiff float64 `json:"monthly_cost_diff"`
	RPS             float64 `json:"rps_used"`
}

// HandlePredictCost handles the /api/v1/finops/predict endpoint.
func HandlePredictCost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PredictCostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	baseBytes := finops.EstimateResponseByteSize(req.BaseSchema)
	proposedBytes := finops.EstimateResponseByteSize(req.ProposedSchema)

	// Use mock Datadog provider interface to get RPS
	provider := finops.NewMockDatadogProvider()
	rps, _ := provider.GetEndpointRPS(req.EndpointPath)

	// Standard cost per GB
	var costPerGB float64 = 0.09

	monthlyCostDiff := finops.CalculateMonthlyCostDiff(baseBytes, proposedBytes, rps, costPerGB)

	resp := PredictCostResponse{
		BaseBytes:       baseBytes,
		ProposedBytes:   proposedBytes,
		MonthlyCostDiff: monthlyCostDiff,
		RPS:             rps,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
