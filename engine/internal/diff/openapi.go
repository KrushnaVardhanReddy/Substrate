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

	// This is a simplified mock adapter for the purpose of the scaffold.
	// In reality, we would map `diffObj` properties to Substrate rules.
	if !diffObj.Empty() {
		// Mock a breaking change if there is any diff at all
		recommendation := "Review changes carefully."
		rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
			ID:             "chg_001",
			RuleID:         "FIELD_REMOVED", // Placeholder mapped rule
			Severity:       report.ChangeSeverityBreaking,
			Path:           "components.schemas",
			Description:    "A difference was detected by oasdiff.",
			Recommendation: &recommendation,
		})

		rep.Summary.TotalChanges = 1
		rep.Summary.BreakingCount = 1
		rep.Summary.OverallSeverity = report.SeverityBreaking
	}

	return rep, nil
}
