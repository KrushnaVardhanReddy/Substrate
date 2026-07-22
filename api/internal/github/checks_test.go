package github

import (
	"context"
	"testing"
)

func TestOrchestrator_CreatePendingChecks(t *testing.T) {
	var createdChecks []string
	mockClient := &MockClient{
		CreatePendingCheckRunFunc: func(ctx context.Context, owner, repo, commitSHA, name, title, summary string) error {
			createdChecks = append(createdChecks, name)
			return nil
		},
	}

	orchestrator := NewOrchestrator(mockClient)
	err := orchestrator.CreatePendingChecks(context.Background(), "owner", "repo", "sha")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(createdChecks) != 4 {
		t.Fatalf("expected 4 pending checks, got %d", len(createdChecks))
	}

	expectedChecks := map[string]bool{
		"substrate / breaking-changes": true,
		"substrate / security-rules":   true,
		"substrate / pii-compliance":   true,
		"substrate / schema-linting":   true,
	}

	for _, check := range createdChecks {
		if !expectedChecks[check] {
			t.Errorf("unexpected check created: %s", check)
		}
	}
}

func TestGetCheckCategory(t *testing.T) {
	tests := []struct {
		ruleID   string
		expected CheckCategory
	}{
		{"FIELD_REMOVED", CheckBreakingChanges},
		{"AUTH_REMOVED", CheckSecurityRules},
		{"SECURITY_RISK", CheckSecurityRules},
		{"PII_ADDED", CheckPIICompliance},
		{"SSN_EXPOSED", CheckPIICompliance},
		{"SCHEMA_LINT_ERROR", CheckSchemaLinting},
		{"DESCRIPTION_MISSING", CheckSchemaLinting},
		{"UNKNOWN_RULE", CheckBreakingChanges},
	}

	for _, tt := range tests {
		t.Run(tt.ruleID, func(t *testing.T) {
			actual := getCheckCategory(tt.ruleID)
			if actual != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, actual)
			}
		})
	}
}

func TestOrchestrator_UpdateChecks(t *testing.T) {
	var createdChecks []struct {
		name       string
		conclusion string
	}
	mockClient := &MockClient{
		CreateCheckRunFunc: func(ctx context.Context, owner, repo, commitSHA, name, title, summary, conclusion string) error {
			createdChecks = append(createdChecks, struct{ name, conclusion string }{name, conclusion})
			return nil
		},
	}

	orchestrator := NewOrchestrator(mockClient)

	anomalies := []DiffAnomaly{
		{RuleID: "FIELD_REMOVED", Description: "Field removed"},
		{RuleID: "SCHEMA_LINT_ERROR", Description: "Missing desc"},
	}

	configFileContent := `
checks:
  schema-linting: advisory
`

	err := orchestrator.UpdateChecks(context.Background(), "owner", "repo", "sha", configFileContent, anomalies)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(createdChecks) != 4 {
		t.Fatalf("expected 4 checks to be updated, got %d", len(createdChecks))
	}

	expectedConclusions := map[string]string{
		"substrate / breaking-changes": "failure", // has FIELD_REMOVED, default blocking
		"substrate / security-rules":   "success", // no anomalies
		"substrate / pii-compliance":   "success", // no anomalies
		"substrate / schema-linting":   "neutral", // has LINT anomaly, but configured as advisory
	}

	for _, check := range createdChecks {
		expected, ok := expectedConclusions[check.name]
		if !ok {
			t.Errorf("unexpected check updated: %s", check.name)
		}
		if check.conclusion != expected {
			t.Errorf("expected %s conclusion to be %s, got %s", check.name, expected, check.conclusion)
		}
	}
}
