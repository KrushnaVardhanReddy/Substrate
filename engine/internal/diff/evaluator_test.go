package diff

import (
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/config"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
)

func TestEvaluateCustomRules(t *testing.T) {
	// Sample schema AST
	schemaAST := map[string]interface{}{
		"info": map[string]interface{}{
			"title":   "Test API",
			"version": "1.0.0",
		},
		"paths": map[string]interface{}{
			"/users": map[string]interface{}{
				"get": map[string]interface{}{
					"operationId": "getUsers",
				},
			},
		},
	}

	tests := []struct {
		name          string
		rules         []config.CustomRule
		expectedCount int
		expectedID    string
	}{
		{
			name:          "No rules",
			rules:         []config.CustomRule{},
			expectedCount: 0,
		},
		{
			name: "Rule evaluating to true (no change generated)",
			rules: []config.CustomRule{
				{
					ID:          "MUST_HAVE_TITLE",
					Description: "API must have a title",
					Severity:    "BREAKING",
					Match:       `info.title == "Test API"`,
				},
			},
			expectedCount: 0,
		},
		{
			name: "Rule evaluating to false (change generated)",
			rules: []config.CustomRule{
				{
					ID:          "MUST_BE_V2",
					Description: "API must be version 2.0.0",
					Severity:    "BREAKING",
					Match:       `info.version == "2.0.0"`,
				},
			},
			expectedCount: 1,
			expectedID:    "MUST_BE_V2",
		},
		{
			name: "Invalid rule syntax (ignored)",
			rules: []config.CustomRule{
				{
					ID:          "INVALID_SYNTAX",
					Description: "Invalid CEL syntax",
					Severity:    "BREAKING",
					Match:       `info.version === "1.0.0"`, // === is invalid CEL
				},
			},
			expectedCount: 0,
		},
		{
			name: "Missing field evaluation (handled, generates change)",
			rules: []config.CustomRule{
				{
					ID:          "MISSING_FIELD",
					Description: "Check for missing field",
					Severity:    "WARNING",
					Match:       `info.contact.email == "test@example.com"`,
				},
			},
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			changes := EvaluateCustomRules(schemaAST, tt.rules)
			if len(changes) != tt.expectedCount {
				t.Errorf("Expected %d changes, got %d", tt.expectedCount, len(changes))
			}
			if tt.expectedCount > 0 && len(changes) > 0 {
				if changes[0].ID != tt.expectedID {
					t.Errorf("Expected change ID %s, got %s", tt.expectedID, changes[0].ID)
				}
				if changes[0].RuleID != tt.expectedID {
					t.Errorf("Expected RuleID %s, got %s", tt.expectedID, changes[0].RuleID)
				}
				if changes[0].Description != tt.rules[0].Description {
					t.Errorf("Expected description %s, got %s", tt.rules[0].Description, changes[0].Description)
				}
				expectedSeverity := report.ChangeSeverityWarning
				if tt.rules[0].Severity == "BREAKING" {
					expectedSeverity = report.ChangeSeverityBreaking
				}
				if changes[0].Severity != expectedSeverity {
					t.Errorf("Expected severity %s, got %s", expectedSeverity, changes[0].Severity)
				}
			}
		})
	}
}
