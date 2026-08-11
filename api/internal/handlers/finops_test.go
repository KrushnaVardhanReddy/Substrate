// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlePredictCost(t *testing.T) {
	reqBody := PredictCostRequest{
		BaseSchema: map[string]interface{}{
			"type": "string",
		},
		ProposedSchema: map[string]interface{}{
			"type": "integer",
		},
		EndpointPath: "/api/v1/test",
	}

	bodyBytes, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/finops/predict", bytes.NewBuffer(bodyBytes))
	w := httptest.NewRecorder()

	HandlePredictCost(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Result().StatusCode)
	}

	var resp PredictCostResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.BaseBytes != 50 {
		t.Errorf("Expected base bytes to be 50, got %d", resp.BaseBytes)
	}
	if resp.ProposedBytes != 8 {
		t.Errorf("Expected proposed bytes to be 8, got %d", resp.ProposedBytes)
	}
	if resp.RPS != 100.0 {
		t.Errorf("Expected RPS to be 100.0, got %f", resp.RPS)
	}
}

func TestHandlePredictCost_WrongMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/finops/predict", nil)
	w := httptest.NewRecorder()

	HandlePredictCost(w, req)

	if w.Result().StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", w.Result().StatusCode)
	}
}

func TestHandlePredictCost_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/finops/predict", bytes.NewBufferString("{invalid json}"))
	w := httptest.NewRecorder()

	HandlePredictCost(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Result().StatusCode)
	}
}
