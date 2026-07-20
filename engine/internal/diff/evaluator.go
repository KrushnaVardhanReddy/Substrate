package diff

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/config"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/traffic"
	"github.com/google/cel-go/cel"
)

// EvaluateCustomRules evaluates a set of CEL rules against the provided schema AST.
// It returns a slice of report.Change for any rules that evaluate to false.
func EvaluateCustomRules(schemaAST interface{}, rules []config.CustomRule) []report.Change {
	var changes []report.Change

	if len(rules) == 0 {
		return changes
	}

	astMap, ok := schemaAST.(map[string]interface{})
	if !ok {
		// If it's not a map, we can't easily inject its keys. We'll still wrap it in "schema".
		astMap = map[string]interface{}{}
	}

	var envOptions []cel.EnvOption
	// Add all top-level keys of the AST as variables, so rules can access them directly.
	for k := range astMap {
		envOptions = append(envOptions, cel.Variable(k, cel.DynType))
	}
	// Also add the entire schema under the variable name "schema".
	envOptions = append(envOptions, cel.Variable("schema", cel.DynType))

	env, err := cel.NewEnv(envOptions...)
	if err != nil {
		// Log or return error? We just skip custom rules on error.
		return changes
	}

	evalCtx := make(map[string]interface{})
	for k, v := range astMap {
		evalCtx[k] = v
	}
	evalCtx["schema"] = schemaAST

	for _, rule := range rules {
		ast, iss := env.Compile(rule.Match)
		if iss.Err() != nil {
			// Skip rules that fail to compile
			continue
		}

		prg, err := env.Program(ast)
		if err != nil {
			continue
		}

		out, _, err := prg.Eval(evalCtx)
		if err != nil {
			// Evaluation failed (e.g., missing field) -> assume rule doesn't match/fails
			continue
		}

		// Ensure it returns a boolean
		res, ok := out.Value().(bool)
		if ok && !res {
			// Rule failed, generate a change
			severity := report.ChangeSeverityWarning // default
			if strings.ToUpper(rule.Severity) == "BREAKING" {
				severity = report.ChangeSeverityBreaking
			} else if strings.ToUpper(rule.Severity) == "SAFE" {
				severity = report.ChangeSeveritySafe
			}

			change := report.Change{
				ID:          rule.ID,
				RuleID:      rule.ID, // Use the custom ID as both ID and RuleID
				Severity:    severity,
				Path:        "custom",
				Description: rule.Description,
			}
			changes = append(changes, change)
		}
	}

	return changes
}

// MapToAST converts any struct to a generic map[string]interface{} (AST) via JSON marshaling.
func MapToAST(obj interface{}) (interface{}, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	var ast interface{}
	err = json.Unmarshal(b, &ast)
	return ast, err
}

// ApplyConfigAndTraffic applies config overrides and traffic-aware downgrades to a DiffReport.
func ApplyConfigAndTraffic(rep *report.DiffReport, cfg *config.SubstrateConfig, providerOrg, providerRepo string) *report.DiffReport {
	if cfg == nil || rep == nil {
		return rep
	}

	var trafficProvider traffic.Provider
	if cfg.Traffic.Provider == "prometheus" {
		trafficProvider = &traffic.PrometheusProvider{Endpoint: cfg.Traffic.Endpoint}
	} else if cfg.Traffic.Provider == "mock" {
		trafficProvider = &traffic.MockProvider{}
	} else {
		trafficProvider = &traffic.NoOpProvider{}
	}

	var activeBreaking []report.Change
	for _, bc := range rep.BreakingChanges {
		if cfg.IsOverrideActive(bc.RuleID, bc.Path) {
			continue
		}

		// Traffic-aware downgrade
		usage, err := trafficProvider.GetFieldUsage(providerOrg, providerRepo, bc.Path, cfg.Traffic.LookbackDays)
		if err == nil && usage >= 0 && usage <= cfg.Traffic.DowngradeThreshold {
			bc.Severity = report.ChangeSeverityWarning
			bc.Description = fmt.Sprintf("%s (Downgraded due to low traffic: %d requests in %d days)", bc.Description, usage, cfg.Traffic.LookbackDays)
			rep.Warnings = append(rep.Warnings, bc)
			rep.Summary.WarningCount++
			continue
		}

		activeBreaking = append(activeBreaking, bc)
	}

	rep.BreakingChanges = activeBreaking
	rep.Summary.BreakingCount = len(rep.BreakingChanges)

	for _, dep := range cfg.Deprecations {
		rep.Deprecations = append(rep.Deprecations, report.Deprecation{
			Endpoint:   dep.Endpoint,
			SunsetDate: dep.SunsetDate,
		})
	}

	if rep.Summary.BreakingCount > 0 {
		rep.Summary.OverallSeverity = report.SeverityBreaking
	} else if rep.Summary.WarningCount > 0 {
		rep.Summary.OverallSeverity = report.SeverityWarning
	} else if rep.Summary.TotalChanges > 0 {
		rep.Summary.OverallSeverity = report.SeveritySafe
	} else {
		rep.Summary.OverallSeverity = report.SeverityNoChanges
	}

	return rep
}
