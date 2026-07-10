package diff_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/diff"
)

func TestCompareTerraformPlan(t *testing.T) {
	tests := []struct {
		name             string
		jsonContent      string
		expectedBreaking int
		expectedWarning  int
		expectedRuleID   string
	}{
		{
			name:             "no destructive changes",
			jsonContent:      `{"resource_changes": [{"address": "aws_s3_bucket.my_bucket", "type": "aws_s3_bucket", "change": {"actions": ["update"]}}]}`,
			expectedBreaking: 0,
			expectedWarning:  0,
		},
		{
			name:             "stateful resource destroyed",
			jsonContent:      `{"resource_changes": [{"address": "aws_dynamodb_table.users", "type": "aws_dynamodb_table", "change": {"actions": ["delete"]}}]}`,
			expectedBreaking: 1,
			expectedWarning:  0,
			expectedRuleID:   "TF_RESOURCE_DESTROYED",
		},
		{
			name:             "output removed",
			jsonContent:      `{"output_changes": {"api_url": {"actions": ["delete"]}}}`,
			expectedBreaking: 1,
			expectedWarning:  0,
			expectedRuleID:   "TF_OUTPUT_REMOVED",
		},
		{
			name:             "iam policy modified",
			jsonContent:      `{"resource_changes": [{"address": "aws_iam_policy.app", "type": "aws_iam_policy", "change": {"actions": ["update"]}}]}`,
			expectedBreaking: 0,
			expectedWarning:  1,
			expectedRuleID:   "TF_IAM_PERMISSION_REMOVED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir := t.TempDir()
			tempFile := filepath.Join(tempDir, "tfplan.json")
			err := os.WriteFile(tempFile, []byte(tt.jsonContent), 0644)
			if err != nil {
				t.Fatalf("failed to write temp file: %v", err)
			}

			rep, err := diff.CompareTerraformPlan(tempFile)
			if err != nil {
				t.Fatalf("CompareTerraformPlan failed: %v", err)
			}

			if rep.Summary.BreakingCount != tt.expectedBreaking {
				t.Errorf("expected %d breaking changes, got %d", tt.expectedBreaking, rep.Summary.BreakingCount)
			}

			if rep.Summary.WarningCount != tt.expectedWarning {
				t.Errorf("expected %d warnings, got %d", tt.expectedWarning, rep.Summary.WarningCount)
			}

			if tt.expectedRuleID != "" {
				found := false
				for _, bc := range rep.BreakingChanges {
					if bc.RuleID == tt.expectedRuleID {
						found = true
						break
					}
				}
				for _, w := range rep.Warnings {
					if w.RuleID == tt.expectedRuleID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected rule ID %q not found in BreakingChanges or Warnings", tt.expectedRuleID)
				}
			}
		})
	}
}
