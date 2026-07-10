package diff

import (
	"fmt"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/config"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/traffic"
)

// ApplyConfigAndTraffic applies config overrides and traffic-aware downgrades to a DiffReport.
func ApplyConfigAndTraffic(rep *report.DiffReport, cfg *config.SubstrateConfig, providerOrg, providerRepo string) *report.DiffReport {
	if cfg == nil || rep == nil {
		return rep
	}

	var trafficProvider traffic.Provider
	if cfg.Traffic.Provider == "prometheus" {
		trafficProvider = &traffic.PrometheusProvider{Endpoint: cfg.Traffic.Endpoint}
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
