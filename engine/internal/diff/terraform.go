package diff

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/compliance"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
)

type TFResourceChange struct {
	Address string `json:"address"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Change  struct {
		Actions []string `json:"actions"`
	} `json:"change"`
}

type TFOutputChange struct {
	Actions []string `json:"actions"`
}

type TFPlan struct {
	ResourceChanges []TFResourceChange        `json:"resource_changes"`
	OutputChanges   map[string]TFOutputChange `json:"output_changes"`
}

var statefulResources = map[string]bool{
	"aws_dynamodb_table":    true,
	"aws_s3_bucket":         true,
	"aws_db_instance":       true,
	"aws_rds_cluster":       true,
	"google_storage_bucket": true,
}

func contains(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func CompareTerraformPlan(planJSONPath string) (*report.DiffReport, error) {
	bytes, err := os.ReadFile(planJSONPath)
	if err != nil {
		return nil, err
	}

	var plan TFPlan
	if err := json.Unmarshal(bytes, &plan); err != nil {
		return nil, err
	}

	rep := &report.DiffReport{
		SchemaType:      report.SchemaType("terraform-plan"),
		ComparedAt:      time.Now().UTC().Format(time.RFC3339),
		BreakingChanges: make([]report.Change, 0),
		Warnings:        make([]report.Change, 0),
		SafeChanges:     make([]report.Change, 0),
	}

	for _, resource := range plan.ResourceChanges {
		actions := resource.Change.Actions

		isDelete := contains(actions, "delete")
		isCreate := contains(actions, "create")
		isUpdate := contains(actions, "update")

		if isDelete && !isCreate {
			if statefulResources[resource.Type] {
				recommendation := "Ensure data is backed up or state is migrated before applying."
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:             "TF_RESOURCE_DESTROYED",
					RuleID:         "TF_RESOURCE_DESTROYED",
					Path:           resource.Address,
					Severity:       report.ChangeSeverityBreaking,
					Description:    fmt.Sprintf("Stateful resource %s is marked for destruction", resource.Address),
					Recommendation: &recommendation,
				})
			}
		} else if isUpdate {
			if resource.Type == "aws_iam_policy" || resource.Type == "aws_iam_role_policy" {
				rep.Warnings = append(rep.Warnings, report.Change{
					ID:          "TF_IAM_PERMISSION_REMOVED",
					RuleID:      "TF_IAM_PERMISSION_REMOVED",
					Path:        resource.Address,
					Severity:    report.ChangeSeverityWarning,
					Description: fmt.Sprintf("IAM policy %s is being modified. Verify no critical permissions are removed.", resource.Address),
				})
			}
		}
	}

	for outputName, outputChange := range plan.OutputChanges {
		actions := outputChange.Actions
		if contains(actions, "delete") {
			rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
				ID:          "TF_OUTPUT_REMOVED",
				RuleID:      "TF_OUTPUT_REMOVED",
				Path:        outputName,
				Severity:    report.ChangeSeverityBreaking,
				Description: fmt.Sprintf("Terraform output '%s' was removed", outputName),
			})
		}
	}

	rep.Summary.TotalChanges = len(rep.BreakingChanges) + len(rep.Warnings) + len(rep.SafeChanges)
	rep.Summary.BreakingCount = len(rep.BreakingChanges)
	rep.Summary.WarningCount = len(rep.Warnings)
	rep.Summary.SafeCount = len(rep.SafeChanges)

	if rep.Summary.BreakingCount > 0 {
		rep.Summary.OverallSeverity = report.SeverityBreaking
	} else if rep.Summary.WarningCount > 0 {
		rep.Summary.OverallSeverity = report.SeverityWarning
	} else if rep.Summary.TotalChanges > 0 {
		rep.Summary.OverallSeverity = report.SeveritySafe
	} else {
		rep.Summary.OverallSeverity = report.SeverityNoChanges
	}

	compliance.Audit(rep)
	return rep, nil
}
