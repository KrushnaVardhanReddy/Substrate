// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package services

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/spf13/viper"

	"github.com/KrushnaVardhanReddy/substrate/engine/pkg/ai"
)

func TestAIImpactHandler(t *testing.T) {
	viper.Set("SUBSTRATE_AI_PROVIDER", "unsupported-to-force-error")
	t.Cleanup(func() {
		viper.Set("SUBSTRATE_AI_PROVIDER", nil)
	})

	handler := AIImpactHandler()

	reqBody := AIImpactRequest{
		Changes: []ai.Change{
			{RuleID: "TEST", Path: "/test", Description: "removed"},
		},
		Consumers: []ai.Repo{
			{Name: "test-repo"},
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/ai/impact", bytes.NewReader(bodyBytes))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp AIImpactResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Explanation == "" {
		t.Errorf("expected non-empty explanation")
	}
}
