// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"testing"
)

func TestGatewaySyncHandler(t *testing.T) {
	handler := GatewaySyncHandler(nil, nil)
	if handler == nil {
		t.Fatal("expected handler to not be nil")
	}
}
