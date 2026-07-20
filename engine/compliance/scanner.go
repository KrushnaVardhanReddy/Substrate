package compliance

import (
	"fmt"
	"regexp"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/config"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
	"github.com/getkin/kin-openapi/openapi3"
)

var defaultPatterns = map[string]*regexp.Regexp{
	"GDPR":  regexp.MustCompile(`(?i)(ssn|social.?security|password|passwd|secret|email|phone|address)`),
	"HIPAA": regexp.MustCompile(`(?i)(medical|diagnosis|hipaa|phi)`),
	"PCI":   regexp.MustCompile(`(?i)(card.?number|pan|cvv|credit_card)`),
	"SOC2":  regexp.MustCompile(`(?i)(token|key|credential|auth)`),
}

func ScanOpenAPISchema(doc *openapi3.T, cfg *config.SubstrateConfig) []report.ComplianceAlert {
	if doc == nil || doc.Components == nil || doc.Components.Schemas == nil {
		return nil
	}

	patterns := make(map[string]*regexp.Regexp)
	for k, v := range defaultPatterns {
		patterns[k] = v
	}

	if cfg != nil && cfg.Compliance != nil && cfg.Compliance.Patterns != nil {
		for tag, patternStr := range cfg.Compliance.Patterns {
			if compiled, err := regexp.Compile(patternStr); err == nil {
				patterns[tag] = compiled
			}
		}
	}

	var alerts []report.ComplianceAlert
	for schemaName, schemaRef := range doc.Components.Schemas {
		visited := make(map[*openapi3.Schema]bool)
		alerts = append(alerts, scanSchema(schemaRef.Value, patterns, "#/components/schemas/"+schemaName, visited)...)
	}
	return alerts
}

func scanSchema(schema *openapi3.Schema, patterns map[string]*regexp.Regexp, currentPath string, visited map[*openapi3.Schema]bool) []report.ComplianceAlert {
	if schema == nil {
		return nil
	}
	if visited[schema] {
		return nil
	}
	visited[schema] = true

	var alerts []report.ComplianceAlert

	if schema.Type.Is("array") && schema.Items != nil && schema.Items.Value != nil {
		alerts = append(alerts, scanSchema(schema.Items.Value, patterns, currentPath+"/items", visited)...)
	}
	for propName, propRef := range schema.Properties {
		if propRef.Value == nil {
			continue
		}

		propPath := currentPath + "/properties/" + propName
		matchedTags := []string{}
		for tag, pattern := range patterns {
			if pattern.MatchString(propName) || pattern.MatchString(propRef.Value.Description) {
				matchedTags = append(matchedTags, tag)

				alerts = append(alerts, report.ComplianceAlert{
					Path:           propPath,
					ComplianceType: tag,
					Message:        fmt.Sprintf("Detected sensitive field matching %s", tag),
				})
			}
		}

		if len(matchedTags) > 0 {
			if propRef.Value.Extensions == nil {
				propRef.Value.Extensions = make(map[string]interface{})
			}
			propRef.Value.Extensions["x-compliance"] = matchedTags
		}

		if propRef.Value.Type.Is("object") {
			alerts = append(alerts, scanSchema(propRef.Value, patterns, propPath, visited)...)
		}
		if propRef.Value.Type.Is("array") && propRef.Value.Items != nil {
			alerts = append(alerts, scanSchema(propRef.Value.Items.Value, patterns, propPath+"/items", visited)...)
		}
	}

	return alerts
}
