package diff

import (
	"fmt"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/diff"
)

// CompareOpenAPI takes two OpenAPI specification paths, computes their diff using oasdiff,
// and maps the result to a Substrate DiffReport.
func CompareOpenAPI(basePath, revisionPath string, flattenAllOf bool) (*report.DiffReport, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	base, err := loader.LoadFromFile(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load base spec: %w", err)
	}

	revision, err := loader.LoadFromFile(revisionPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load revision spec: %w", err)
	}

	// Create diff config
	config := diff.NewConfig()

	// Compute the diff using oasdiff
	diffObj, err := diff.Get(config, base, revision)
	if err != nil {
		return nil, fmt.Errorf("failed to compute diff: %w", err)
	}

	rep := &report.DiffReport{
		SubstrateVersion: "0.1.0",
		SchemaType:       report.SchemaTypeOpenAPI,
		ComparedAt:       time.Now().UTC().Format(time.RFC3339),
		Summary: report.Summary{
			TotalChanges:    0,
			BreakingCount:   0,
			WarningCount:    0,
			SafeCount:       0,
			OverallSeverity: report.SeverityNoChanges,
		},
		BreakingChanges: []report.Change{},
		Warnings:        []report.Change{},
		SafeChanges:     []report.Change{},
	}

	if !diffObj.Empty() {
		if diffObj.PathsDiff != nil {
			for _, pathName := range diffObj.PathsDiff.Deleted {
				recommendation := "Add 'deprecated: true' before removing endpoints."
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:             fmt.Sprintf("chg_path_del_%v", pathName),
					RuleID:         "ENDPOINT_REMOVED",
					Severity:       report.ChangeSeverityBreaking,
					Path:           fmt.Sprintf("paths.%v", pathName),
					Description:    fmt.Sprintf("Endpoint %v was removed.", pathName),
					Recommendation: &recommendation,
				})
			}

			// Example: Check for modified paths
			for _, pathName := range diffObj.PathsDiff.Modified {
				recommendation := "Review path modifications carefully."
				rep.Warnings = append(rep.Warnings, report.Change{
					ID:             fmt.Sprintf("chg_path_mod_%v", pathName),
					RuleID:         "ENDPOINT_MODIFIED",
					Severity:       report.ChangeSeverityWarning,
					Path:           fmt.Sprintf("paths.%v", pathName),
					Description:    fmt.Sprintf("Endpoint %v was modified.", pathName),
					Recommendation: &recommendation,
				})
			}
		}

		if diffObj.ComponentsDiff != nil && diffObj.ComponentsDiff.SchemasDiff != nil {
			for schemaName := range diffObj.ComponentsDiff.SchemasDiff.Deleted {
				recommendation := "Avoid deleting schemas in use."
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:             fmt.Sprintf("chg_schema_del_%v", schemaName),
					RuleID:         "SCHEMA_REMOVED",
					Severity:       report.ChangeSeverityBreaking,
					Path:           fmt.Sprintf("components.schemas.%v", schemaName),
					Description:    fmt.Sprintf("Schema %v was removed.", schemaName),
					Recommendation: &recommendation,
				})
			}

			for schemaName := range diffObj.ComponentsDiff.SchemasDiff.Modified {
				recommendation := "Modifying schemas can cause breaking changes."
				rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
					ID:             fmt.Sprintf("chg_schema_mod_%v", schemaName),
					RuleID:         "FIELD_REMOVED",
					Severity:       report.ChangeSeverityBreaking,
					Path:           fmt.Sprintf("components.schemas.%v", schemaName),
					Description:    fmt.Sprintf("Schema %v was modified.", schemaName),
					Recommendation: &recommendation,
				})
			}
		}
	}

	rep.Summary.BreakingCount = len(rep.BreakingChanges)
	rep.Summary.WarningCount = len(rep.Warnings)
	rep.Summary.SafeCount = len(rep.SafeChanges)
	rep.Summary.TotalChanges = rep.Summary.BreakingCount + rep.Summary.WarningCount + rep.Summary.SafeCount

	if rep.Summary.BreakingCount > 0 {
		rep.Summary.OverallSeverity = report.SeverityBreaking
	} else if rep.Summary.WarningCount > 0 {
		rep.Summary.OverallSeverity = report.SeverityWarning
	} else if rep.Summary.SafeCount > 0 {
		rep.Summary.OverallSeverity = report.SeveritySafe
	} else {
		rep.Summary.OverallSeverity = report.SeverityNoChanges
	}

	return rep, nil
}
