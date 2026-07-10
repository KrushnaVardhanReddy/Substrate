package compliance

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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

// compiledPatterns caches the compiled regexes
var compiledPatterns map[*regexp.Regexp]string

func init() {
	compiledPatterns = make(map[*regexp.Regexp]string)
	for pattern, tag := range piiPatterns {
		compiledPatterns[regexp.MustCompile(pattern)] = tag
	}
}

// Audit scans a DiffReport for compliance risks and sends alerts if configured
func Audit(rep *report.DiffReport) {
	alerts := make([]report.ComplianceAlert, 0)

	// Scan through all changes (added fields might be breaking or safe depending on context,
	// but generally added response fields or request fields are safe or warnings).
	// We'll scan the Path string which usually contains the field name.
	checkChanges := func(changes []report.Change) {
		for _, c := range changes {
			// We only care about new fields, but checking everything is fine for MVP
			for regex, tag := range compiledPatterns {
				if regex.MatchString(c.Path) {
					alerts = append(alerts, report.ComplianceAlert{
						FieldPath:     c.Path,
						ComplianceTag: tag,
						Reason:        fmt.Sprintf("Path matches %s pattern", tag),
					})
				}
			}
		}
	}

	checkChanges(rep.BreakingChanges)
	checkChanges(rep.Warnings)
	checkChanges(rep.SafeChanges)

	if len(alerts) > 0 {
		rep.ComplianceAlerts = alerts
		notifySecurityTeam(alerts)
	}
}

func notifySecurityTeam(alerts []report.ComplianceAlert) {
	webhookURL := os.Getenv("SUBSTRATE_SLACK_WEBHOOK")
	if webhookURL == "" {
		return // Not configured
	}

	channel := os.Getenv("SUBSTRATE_SECURITY_CHANNEL")
	if channel == "" {
		channel = "#security-alerts"
	}

	text := "🚨 *Substrate Compliance Alert*\nFound sensitive fields in recent schema changes:\n"
	for _, a := range alerts {
		text += fmt.Sprintf("• `%s` matched tag `%s`\n", a.FieldPath, a.ComplianceTag)
	}

	payload := map[string]string{
		"channel": channel,
		"text":    text,
	}
	body, _ := json.Marshal(payload)

	http.Post(webhookURL, "application/json", bytes.NewBuffer(body))
}
