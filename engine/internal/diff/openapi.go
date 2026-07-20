package diff

import (
	"fmt"
	"strings"
	"sort"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/compliance"
	extcompliance "github.com/KrushnaVardhanReddy/substrate/engine/compliance"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/config"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/oasdiff/oasdiff/checker"
	"github.com/oasdiff/oasdiff/diff"
)

// CompareOpenAPI takes two OpenAPI specification paths, computes their diff using oasdiff,
// and maps the result to a Substrate DiffReport.
func CompareOpenAPI(basePath, revisionPath string, flattenAllOf bool, customRules []config.CustomRule) (*report.DiffReport, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true

	// Step 1: Load specs
	base, err := loader.LoadFromFile(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load base spec: %w", err)
	}

	revision, err := loader.LoadFromFile(revisionPath)
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

	alerts1 := extcompliance.ScanOpenAPISchema(base, nil)
	alerts2 := extcompliance.ScanOpenAPISchema(revision, nil)


	// Step 3: Compute structural diff using diff.Get()
	// flattenAllOf is accepted for API compatibility but is not currently wired —
	// oasdiff v1.22.0 diff.Config does not expose a FlattenAllOf option.
	// oasdiff handles allOf internally. Revisit if a future version adds this.
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
	// Merge unique compliance alerts
	alertMap := make(map[string]report.ComplianceAlert)
	for _, a := range append(alerts1, alerts2...) {
		alertMap[a.Path+a.ComplianceType] = a
	}
	for _, a := range alertMap {
		rep.ComplianceAlerts = append(rep.ComplianceAlerts, a)
	}

	sort.Slice(rep.ComplianceAlerts, func(i, j int) bool {
		if rep.ComplianceAlerts[i].Path == rep.ComplianceAlerts[j].Path {
			return rep.ComplianceAlerts[i].ComplianceType < rep.ComplianceAlerts[j].ComplianceType
		}
		return rep.ComplianceAlerts[i].Path < rep.ComplianceAlerts[j].Path
	})

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

func mapOasdiffRule(oasdiffID string) (string, report.ChangeSeverity) {
	// Mappings based on docs/specs/breaking-change-rules.md
	switch oasdiffID {
	case "response-property-removed", "request-property-removed", "response-optional-property-removed", "request-optional-property-removed", "response-required-property-removed", "request-required-property-removed":
		return "FIELD_REMOVED", report.ChangeSeverityBreaking
	case "response-property-added", "response-optional-property-added":
		return "FIELD_ADDED_OPTIONAL", report.ChangeSeveritySafe
	case "new-required-request-property":
		return "REQUIRED_FIELD_ADDED", report.ChangeSeverityBreaking
	case "api-path-removed-without-deprecation", "api-path-removed":
		return "ENDPOINT_REMOVED", report.ChangeSeverityBreaking
	case "api-removed-without-deprecation", "api-removed":
		return "METHOD_REMOVED", report.ChangeSeverityBreaking
	case "api-path-added", "endpoint-added":
		return "ENDPOINT_ADDED", report.ChangeSeveritySafe
	case "api-added":
		return "METHOD_ADDED", report.ChangeSeveritySafe
	case "api-deprecated", "endpoint-deprecated", "api-deprecated-without-sunset":
		return "ENDPOINT_DEPRECATED", report.ChangeSeverityWarning
	case "request-parameter-enum-value-removed", "request-property-enum-value-removed", "response-property-enum-value-removed":
		return "ENUM_VALUE_REMOVED", report.ChangeSeverityBreaking
	case "request-parameter-enum-value-added", "request-property-enum-value-added", "response-property-enum-value-added":
		return "ENUM_VALUE_ADDED", report.ChangeSeverityWarning
	case "request-parameter-became-required", "request-property-became-required", "response-property-became-required":
		return "PARAMETER_MADE_REQUIRED", report.ChangeSeverityBreaking
	case "request-parameter-type-changed":
		return "PARAMETER_TYPE_CHANGED", report.ChangeSeverityBreaking
	case "response-property-type-changed", "request-property-type-changed":
		return "FIELD_TYPE_CHANGED", report.ChangeSeverityBreaking
	case "request-parameter-removed":
		return "PARAMETER_REMOVED", report.ChangeSeverityBreaking
	case "request-parameter-added":
		return "PARAMETER_ADDED_OPTIONAL", report.ChangeSeveritySafe
	case "new-required-request-parameter":
		return "PARAMETER_ADDED_REQUIRED", report.ChangeSeverityBreaking
	case "response-required-property-added":
		return "RESPONSE_SCHEMA_FIELD_ADDED", report.ChangeSeveritySafe
	case "response-property-became-optional":
		return "REQUIRED_FIELD_MADE_OPTIONAL", report.ChangeSeveritySafe
	default:
		// Unmapped oasdiff rule IDs must be included in the report with a generated
		// Substrate ID: "OASDIFF_" + strings.ToUpper(oasdiffID)
		ruleID := "OASDIFF_" + strings.ToUpper(oasdiffID)
		return ruleID, report.ChangeSeverityWarning // default to warning if unknown
	}
}
