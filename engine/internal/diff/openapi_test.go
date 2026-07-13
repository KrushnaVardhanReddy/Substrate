package diff

import (
	"strings"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
)

func TestCompareOpenAPI(t *testing.T) {
	tests := []struct {
		name           string
		baseFile       string
		revisionFile   string
		expectedSev    report.Severity
		expectedRuleID string
		expectError    bool
		expectedBucket string // "breaking", "warning", "safe", or "none"
	}{
		{
			name:           "No changes (identical specs)",
			baseFile:       "testdata/base_no_change.yaml",
			revisionFile:   "testdata/rev_no_change.yaml",
			expectedSev:    report.SeverityNoChanges,
			expectedBucket: "none",
		},
		{
			name:           "Optional field added -> SAFE / FIELD_ADDED_OPTIONAL",
			baseFile:       "testdata/base_optional_add.yaml",
			revisionFile:   "testdata/rev_optional_add.yaml",
			expectedSev:    report.SeveritySafe,
			expectedRuleID: "FIELD_ADDED_OPTIONAL",
			expectedBucket: "safe",
		},
		{
			name:           "Required field added -> BREAKING / REQUIRED_FIELD_ADDED",
			baseFile:       "testdata/base_required_add.yaml",
			revisionFile:   "testdata/rev_required_add.yaml",
			expectedSev:    report.SeverityBreaking,
			expectedRuleID: "REQUIRED_FIELD_ADDED",
			expectedBucket: "breaking",
		},
		{
			name:           "Field removed -> BREAKING / FIELD_REMOVED",
			baseFile:       "testdata/base_field_removed.yaml",
			revisionFile:   "testdata/rev_field_removed.yaml",
			expectedSev:    report.SeverityBreaking,
			expectedRuleID: "FIELD_REMOVED",
			expectedBucket: "breaking",
		},
		{
			name:           "Endpoint removed -> BREAKING / ENDPOINT_REMOVED",
			baseFile:       "testdata/base_endpoint_removed.yaml",
			revisionFile:   "testdata/rev_endpoint_removed.yaml",
			expectedSev:    report.SeverityBreaking,
			expectedRuleID: "ENDPOINT_REMOVED",
			expectedBucket: "breaking",
		},
		{
			name:           "Method removed -> BREAKING / METHOD_REMOVED",
			baseFile:       "testdata/base_method_removed.yaml",
			revisionFile:   "testdata/rev_method_removed.yaml",
			expectedSev:    report.SeverityBreaking,
			expectedRuleID: "METHOD_REMOVED",
			expectedBucket: "breaking",
		},
		{
			name:           "Enum value removed -> BREAKING / ENUM_VALUE_REMOVED",
			baseFile:       "testdata/base_enum_removed.yaml",
			revisionFile:   "testdata/rev_enum_removed.yaml",
			expectedSev:    report.SeverityBreaking,
			expectedRuleID: "ENUM_VALUE_REMOVED",
			expectedBucket: "breaking",
		},
		{
			name:           "Enum value added -> WARNING / ENUM_VALUE_ADDED",
			baseFile:       "testdata/base_enum_added.yaml",
			revisionFile:   "testdata/rev_enum_added.yaml",
			expectedSev:    report.SeverityWarning,
			expectedRuleID: "ENUM_VALUE_ADDED",
			expectedBucket: "warning",
		},
		{
			name:           "Parameter made required -> BREAKING / PARAMETER_MADE_REQUIRED",
			baseFile:       "testdata/base_param_required.yaml",
			revisionFile:   "testdata/rev_param_required.yaml",
			expectedSev:    report.SeverityBreaking,
			expectedRuleID: "PARAMETER_MADE_REQUIRED",
			expectedBucket: "breaking",
		},
		{
			name:           "Type changed -> BREAKING / FIELD_TYPE_CHANGED",
			baseFile:       "testdata/base_type_changed.yaml",
			revisionFile:   "testdata/rev_type_changed.yaml",
			expectedSev:    report.SeverityBreaking,
			expectedRuleID: "FIELD_TYPE_CHANGED",
			expectedBucket: "breaking",
		},
		{
			name:         "Invalid base spec -> error returned, no DiffReport",
			baseFile:     "testdata/base_invalid.yaml",
			revisionFile: "testdata/base_invalid.yaml", // using same invalid for both is fine
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rep, err := CompareOpenAPI(tt.baseFile, tt.revisionFile, true, nil)
			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error for invalid spec, but got nil")
				}
				return // Success on expecting error
			}

			if err != nil {
				t.Fatalf("CompareOpenAPI failed: %v", err)
			}

			if rep.Summary.OverallSeverity != tt.expectedSev {
				t.Errorf("Expected severity %s, got %s", tt.expectedSev, rep.Summary.OverallSeverity)
			}

			// Validate bucket counts based on expectedBucket
			breakingCount := len(rep.BreakingChanges)
			warningCount := len(rep.Warnings)
			safeCount := len(rep.SafeChanges)

			if rep.Summary.BreakingCount != breakingCount {
				t.Errorf("BreakingCount %d does not match length of array %d", rep.Summary.BreakingCount, breakingCount)
			}
			if rep.Summary.WarningCount != warningCount {
				t.Errorf("WarningCount %d does not match length of array %d", rep.Summary.WarningCount, warningCount)
			}
			if rep.Summary.SafeCount != safeCount {
				t.Errorf("SafeCount %d does not match length of array %d", rep.Summary.SafeCount, safeCount)
			}

			if tt.expectedBucket == "none" {
				if breakingCount > 0 || warningCount > 0 || safeCount > 0 {
					t.Errorf("Expected no changes, got B:%d W:%d S:%d", breakingCount, warningCount, safeCount)
				}
				return
			}

			var found bool
			var changes []report.Change

			switch tt.expectedBucket {
			case "breaking":
				if breakingCount == 0 {
					t.Errorf("Expected breaking changes, got none")
				}
				changes = rep.BreakingChanges
			case "warning":
				if warningCount == 0 {
					t.Errorf("Expected warning changes, got none")
				}
				changes = rep.Warnings
			case "safe":
				if safeCount == 0 {
					t.Errorf("Expected safe changes, got none")
				}
				changes = rep.SafeChanges
			}

			if tt.expectedRuleID != "" {
				for _, c := range changes {
					if c.RuleID == tt.expectedRuleID {
						found = true
						break
					}
				}
				if !found {
					// Fallback to see if it's there at all
					var ids []string
					for _, c := range changes {
						ids = append(ids, c.RuleID)
					}
					t.Errorf("Expected RuleID %s not found in %s bucket. Found IDs: %v", tt.expectedRuleID, tt.expectedBucket, strings.Join(ids, ", "))
				}
			}
		})
	}
}
