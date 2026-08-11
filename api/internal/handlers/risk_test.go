// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRiskScoreHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/risk/testorg/testrepo/123", nil)
	req.SetPathValue("org", "testorg")
	req.SetPathValue("repo", "testrepo")
	req.SetPathValue("pr", "123")

	w := httptest.NewRecorder()
	handler := RiskScoreHandler()
	handler.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %v", res.StatusCode)
	}

	var response map[string]string
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["score"] != "LOW" {
		t.Errorf("Expected score LOW, got %v", response["score"])
	}
}

func TestRiskScoreHandler_High(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/risk/testorg/testrepo/123?blast_radius_node_count=6", nil)
	req.SetPathValue("org", "testorg")
	req.SetPathValue("repo", "testrepo")
	req.SetPathValue("pr", "123")

	w := httptest.NewRecorder()
	handler := RiskScoreHandler()
	handler.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %v", res.StatusCode)
	}

	var response map[string]string
	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if response["score"] != "HIGH" {
		t.Errorf("Expected score HIGH, got %v", response["score"])
	}
}

func TestRiskScoreHandler_MissingParams(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/v1/risk///123", nil)
	req.SetPathValue("org", "")
	req.SetPathValue("repo", "")
	req.SetPathValue("pr", "123")

	w := httptest.NewRecorder()
	handler := RiskScoreHandler()
	handler.ServeHTTP(w, req)

	res := w.Result()
	if res.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %v", res.StatusCode)
	}
}
