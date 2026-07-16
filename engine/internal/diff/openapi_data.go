package diff

import (
	"fmt"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/compliance"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/config"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
)

// CompareOpenAPIFromData compares two OpenAPI specifications given as byte arrays
func CompareOpenAPIFromData(baseData, revisionData []byte, flattenAllOf bool, customRules []config.CustomRule) (*report.DiffReport, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// Step 1: Load specs
	base, err := loader.LoadFromData(baseData)
	if err != nil {
		return nil, fmt.Errorf("failed to load base spec: %w", err)
	}

	revision, err := loader.LoadFromData(revisionData)
	if err != nil {
		return nil, fmt.Errorf("failed to load revision spec: %w", err)
	}

	// Step 2: Pre-diff spec validation using openapi3.Validate()
	if err := base.Validate(loader.Context); err != nil {
		return nil, fmt.Errorf("invalid base spec: %w", err)
	}
	if err := revision.Validate(loader.Context); err != nil {
		return nil, fmt.Errorf("invalid revision spec: %w", err)
	}

	// Step 3: Compute structural diff using diff.Get()
	_ = flattenAllOf
	diffConfig := diff.NewConfig()

	diffObj, err := diff.Get(diffConfig, base, revision)
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

	if diffObj.Empty() {
		compliance.Audit(rep)
		return rep, nil
	}

	// Step 4: Run checker.CheckBackwardCompatibility()
	checkerConfig := checker.NewConfig(checker.GetAllChecks())
	osMap := &diff.OperationsSourcesMap{}
	changes := checker.CheckBackwardCompatibilityUntilLevel(checkerConfig, diffObj, osMap, checker.INFO)

	// Step 5: Map checker.BackwardCompatibilityErrors -> report.Change objects
	l := checker.NewLocalizer("en")
	for _, c := range changes {
		substrateID, severity := mapOasdiffRule(c.GetId())

		description := c.GetText(l)
		// Fallback if empty
		if description == "" {
			description = fmt.Sprintf("Change detected: %s", c.GetId())
		}

		changePath := c.GetPath()
		if changePath == "" {
			changePath = "unknown"
		}
		if c.GetOperation() != "" && c.GetPath() != "" {
			changePath = fmt.Sprintf("%s %s", c.GetOperation(), c.GetPath())
		}

		changeObj := report.Change{
			ID:          c.GetId(),
			RuleID:      substrateID,
			Severity:    severity,
			Path:        changePath,
			Description: description,
		}

		switch severity {
		case report.ChangeSeverityBreaking:
			rep.BreakingChanges = append(rep.BreakingChanges, changeObj)
		case report.ChangeSeverityWarning:
			rep.Warnings = append(rep.Warnings, changeObj)
		case report.ChangeSeveritySafe:
			rep.SafeChanges = append(rep.SafeChanges, changeObj)
		}
	}

	// Evaluate custom CEL rules on the head/revision schema
	if len(customRules) > 0 {
		ast, err := MapToAST(revision)
		if err == nil {
			customRuleChanges := EvaluateCustomRules(ast, customRules)
			for _, c := range customRuleChanges {
				switch c.Severity {
				case report.ChangeSeverityBreaking:
					rep.BreakingChanges = append(rep.BreakingChanges, c)
				case report.ChangeSeverityWarning:
					rep.Warnings = append(rep.Warnings, c)
				case report.ChangeSeveritySafe:
					rep.SafeChanges = append(rep.SafeChanges, c)
				}
			}
		}
	}

	// Step 6: Assemble and return the DiffReport
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

	compliance.Audit(rep)
	return rep, nil
}
