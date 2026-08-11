// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package report

import (
	"encoding/json"
	"testing"
)

func TestDiffReportJSON(t *testing.T) {
	recommendation1 := "Add 'deprecated: true' to the field for one release cycle before removing it. Coordinate with consumers: billing-service, marketing-service."
	recommendation2 := "Consumers using exhaustive switch/match statements on this enum may fail on the new value. Verify downstream handlers."

	report := DiffReport{
		SubstrateVersion: "0.1.0",
		SchemaType:       SchemaTypeOpenAPI,
		ComparedAt:       "2026-07-07T01:00:00Z",
		Summary: Summary{
			TotalChanges:    3,
			BreakingCount:   1,
			WarningCount:    1,
			SafeCount:       1,
			OverallSeverity: SeverityBreaking,
		},
		BreakingChanges: []Change{
			{
				ID:          "chg_001",
				RuleID:      "FIELD_REMOVED",
				Severity:    ChangeSeverityBreaking,
				Path:        "components.schemas.Customer.properties.email",
				Description: "Field 'email' was removed from schema 'Customer'.",
				Before: map[string]interface{}{
					"type":   "string",
					"format": "email",
				},
				After:          nil,
				Recommendation: &recommendation1,
			},
		},
		Warnings: []Change{
			{
				ID:             "chg_002",
				RuleID:         "ENUM_VALUE_ADDED",
				Severity:       ChangeSeverityWarning,
				Path:           "components.schemas.OrderStatus.enum",
				Description:    "New enum value 'DISPUTED' was added to 'OrderStatus'.",
				Before:         []interface{}{"PENDING", "COMPLETED", "CANCELLED"},
				After:          []interface{}{"PENDING", "COMPLETED", "CANCELLED", "DISPUTED"},
				Recommendation: &recommendation2,
			},
		},
		SafeChanges: []Change{
			{
				ID:          "chg_003",
				RuleID:      "OPTIONAL_FIELD_ADDED",
				Severity:    ChangeSeveritySafe,
				Path:        "components.schemas.Customer.properties.phone",
				Description: "Optional field 'phone' was added to schema 'Customer'.",
				Before:      nil,
				After: map[string]interface{}{
					"type":     "string",
					"nullable": true,
				},
				Recommendation: nil,
			},
		},
	}

	// Test Marshal
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("Failed to marshal DiffReport: %v", err)
	}

	// Test Unmarshal
	var unmarshaledReport DiffReport
	err = json.Unmarshal(data, &unmarshaledReport)
	if err != nil {
		t.Fatalf("Failed to unmarshal DiffReport: %v", err)
	}

	// Verify fields
	if unmarshaledReport.SubstrateVersion != report.SubstrateVersion {
		t.Errorf("Expected SubstrateVersion %q, got %q", report.SubstrateVersion, unmarshaledReport.SubstrateVersion)
	}
	if len(unmarshaledReport.BreakingChanges) != 1 {
		t.Errorf("Expected 1 breaking change, got %d", len(unmarshaledReport.BreakingChanges))
	}
	if *unmarshaledReport.BreakingChanges[0].Recommendation != recommendation1 {
		t.Errorf("Expected recommendation %q, got %q", recommendation1, *unmarshaledReport.BreakingChanges[0].Recommendation)
	}
}
