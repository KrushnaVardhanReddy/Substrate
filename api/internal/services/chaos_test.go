// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package services

import (
	"strings"
	"testing"
)

func TestChaosTestingService_Mock(t *testing.T) {
	service := NewChaosTestingService()
	// Force mock mode
	service.aiConfig.BaseURL = ""

	code, err := service.GenerateChaosTest("old", "new", "break")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !strings.Contains(code, "assert.ok(newResponse.user_id") {
		t.Errorf("expected generated code to contain assertion on user_id")
	}
}
