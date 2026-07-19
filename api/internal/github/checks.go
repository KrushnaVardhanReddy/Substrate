package github

import (
	"context"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// CheckCategory represents the different types of checks we run
type CheckCategory string

const (
	CheckBreakingChanges CheckCategory = "substrate / breaking-changes"
	CheckSecurityRules   CheckCategory = "substrate / security-rules"
	CheckPIICompliance   CheckCategory = "substrate / pii-compliance"
	CheckSchemaLinting   CheckCategory = "substrate / schema-linting"
)

// Config represents the substrate.yaml config structure for checks
type Config struct {
	Checks map[string]string `yaml:"checks"`
}

// DiffAnomaly represents an anomaly from the Diff Engine
type DiffAnomaly struct {
	RuleID      string
	Description string
}

// Orchestrator manages check runs for a PR
type Orchestrator struct {
	client Client
}

// NewOrchestrator creates a new Orchestrator
func NewOrchestrator(client Client) *Orchestrator {
	return &Orchestrator{client: client}
}

// CreatePendingChecks creates the 4 pending checks when a PR starts
func (o *Orchestrator) CreatePendingChecks(ctx context.Context, owner, repo, commitSHA string) error {
	checks := []CheckCategory{
		CheckBreakingChanges,
		CheckSecurityRules,
		CheckPIICompliance,
		CheckSchemaLinting,
	}

	for _, check := range checks {
		err := o.client.CreatePendingCheckRun(
			ctx,
			owner,
			repo,
			commitSHA,
			string(check),
			"Substrate Analysis Pending",
			"Waiting for Diff Engine analysis to complete.",
		)
		if err != nil {
			return fmt.Errorf("failed to create pending check %s: %w", check, err)
		}
	}
	return nil
}

// getCheckCategory maps a rule_id to a CheckCategory
func getCheckCategory(ruleID string) CheckCategory {
	switch {
	case strings.Contains(strings.ToUpper(ruleID), "AUTH_REMOVED") || strings.Contains(strings.ToUpper(ruleID), "SECURITY"):
		return CheckSecurityRules
	case strings.Contains(strings.ToUpper(ruleID), "PII") || strings.Contains(strings.ToUpper(ruleID), "SSN"):
		return CheckPIICompliance
	case strings.Contains(strings.ToUpper(ruleID), "LINT") || strings.Contains(strings.ToUpper(ruleID), "DESCRIPTION_MISSING"):
		return CheckSchemaLinting
	default:
		return CheckBreakingChanges // e.g. FIELD_REMOVED goes here
	}
}

// Conclusion represents a GitHub check conclusion
type Conclusion string

const (
	ConclusionSuccess Conclusion = "success"
	ConclusionFailure Conclusion = "failure"
	ConclusionNeutral Conclusion = "neutral"
)

// UpdateChecks evaluates anomalies against configuration and updates check runs
func (o *Orchestrator) UpdateChecks(ctx context.Context, owner, repo, commitSHA, configFileContent string, anomalies []DiffAnomaly) error {
	var cfg Config
	if configFileContent != "" {
		if err := yaml.Unmarshal([]byte(configFileContent), &cfg); err != nil {
			// ignore parse errors and proceed with default config
		}
	}

	// Group anomalies by category
	grouped := make(map[CheckCategory][]DiffAnomaly)
	for _, a := range anomalies {
		cat := getCheckCategory(a.RuleID)
		grouped[cat] = append(grouped[cat], a)
	}

	checks := []CheckCategory{
		CheckBreakingChanges,
		CheckSecurityRules,
		CheckPIICompliance,
		CheckSchemaLinting,
	}

	for _, check := range checks {
		anoms := grouped[check]

		// Map CheckCategory name to config key
		configKey := ""
		switch check {
		case CheckBreakingChanges:
			configKey = "breaking-changes"
		case CheckSecurityRules:
			configKey = "security-rules"
		case CheckPIICompliance:
			configKey = "pii-compliance"
		case CheckSchemaLinting:
			configKey = "schema-linting"
		}

		enforcementLevel := "blocking"
		if cfg.Checks != nil {
			if val, ok := cfg.Checks[configKey]; ok {
				enforcementLevel = strings.ToLower(val)
			}
		}

		conclusion := ConclusionSuccess
		summary := "All checks passed."
		title := "Success"

		if len(anoms) > 0 {
			if enforcementLevel == "advisory" {
				conclusion = ConclusionNeutral
			} else {
				conclusion = ConclusionFailure
			}

			title = fmt.Sprintf("%d issues found", len(anoms))
			var descriptions []string
			for _, a := range anoms {
				descriptions = append(descriptions, fmt.Sprintf("- [%s] %s", a.RuleID, a.Description))
			}
			summary = "The following issues were detected:\n" + strings.Join(descriptions, "\n")
		}

		// GitHub App Update API is essentially CreateCheckRun with updated status and conclusion.
		// For our client, CreateCheckRun handles setting the final status to completed.
		// Wait, the signature of CreateCheckRun in our mock/interface just sets conclusion = "success" hardcoded? Let's check client.go

		err := o.client.CreateCheckRun(ctx, owner, repo, commitSHA, string(check), title, summary, string(conclusion))
		if err != nil {
			return fmt.Errorf("failed to update check %s: %w", check, err)
		}
	}

	return nil
}
