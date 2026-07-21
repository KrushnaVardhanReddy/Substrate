package compliance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"net/http"
	"regexp"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
)

var piiPatterns = map[string]string{
	`(?i)(ssn|social.?security)`:    "PII:SSN",
	`(?i)(password|passwd|secret)`:  "SECURITY:CREDENTIAL",
	`(?i)(card.?number|pan|cvv)`:    "PCI:PAYMENT",
	`(?i)(medical|diagnosis|hipaa)`: "HIPAA:PHI",
	`(?i)(email|phone|address)`:     "PII:CONTACT",
}

// Compile regexes once
var compiledPatterns map[string]*regexp.Regexp

func init() {
	compiledPatterns = make(map[string]*regexp.Regexp)
	for pattern, complianceType := range piiPatterns {
		compiledPatterns[complianceType] = regexp.MustCompile(pattern)
	}
}

func Audit(rep *report.DiffReport) {
	if rep == nil {
		return
	}

	var allChanges []report.Change
	allChanges = append(allChanges, rep.BreakingChanges...)
	allChanges = append(allChanges, rep.Warnings...)
	allChanges = append(allChanges, rep.SafeChanges...)

	for _, change := range allChanges {
		for complianceType, regex := range compiledPatterns {
			if regex.MatchString(change.Path) || regex.MatchString(change.Description) {
				alert := report.ComplianceAlert{
					Path:           change.Path,
					ComplianceType: complianceType,
					Message:        fmt.Sprintf("Detected field matching %s pattern", complianceType),
				}
				rep.ComplianceAlerts = append(rep.ComplianceAlerts, alert)

				notifySecurityTeam(alert)
			}
		}
	}
}

func notifySecurityTeam(alert report.ComplianceAlert) {
	webhookURL := viper.GetString("SUBSTRATE_SLACK_WEBHOOK")
	if webhookURL == "" {
		return
	}

	payload := map[string]interface{}{
		"text": fmt.Sprintf("🚨 *Compliance Alert*: %s detected at `%s`\n%s", alert.ComplianceType, alert.Path, alert.Message),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return
	}

	_, _ = http.Post(webhookURL, "application/json", bytes.NewBuffer(jsonData))
}
