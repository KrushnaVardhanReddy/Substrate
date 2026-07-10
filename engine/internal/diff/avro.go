package diff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/compliance"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/config"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/report"
)

type schemaRegistryCompatResponse struct {
	IsCompatible bool     `json:"is_compatible"`
	Messages     []string `json:"messages"`
}

func CompareAvro(baseFile, headFile string, cfg *config.SubstrateConfig) (*report.DiffReport, error) {
	rep := &report.DiffReport{
		SchemaType: "avro",
		ComparedAt: time.Now().UTC().Format(time.RFC3339),
	}

	if cfg == nil || cfg.Avro == nil || cfg.Avro.SchemaRegistryURL == "" {
		rep.Warnings = append(rep.Warnings, report.Change{
			ID:             "avro_no_registry_configured",
			RuleID:         "AVRO_NO_REGISTRY_CONFIGURED",
			Severity:       report.ChangeSeverity(report.SeverityWarning),
			Description:    "Avro compatibility checking requires schema_registry_url in substrate.yaml under the avro: block",
			Recommendation: func(s string) *string { return &s }("Add avro:\n  schema_registry_url: https://your-registry.com\n  subject: your-topic-value"),
		})
		rep.Summary.WarningCount = 1
		compliance.Audit(rep)
		return rep, nil
	}

	headBytes, err := os.ReadFile(headFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read head avro schema: %w", err)
	}

	subject := cfg.Service + "-value"
	if cfg.Service == "" {
		subject = "unknown-value"
	}
	if cfg.Avro.Subject != "" {
		subject = cfg.Avro.Subject
	}

	url := fmt.Sprintf("%s/compatibility/subjects/%s/versions/latest", cfg.Avro.SchemaRegistryURL, subject)

	reqBody, err := json.Marshal(map[string]string{
		"schema": string(headBytes),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal avro schema for registry: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/vnd.schemaregistry.v1+json")
	req.Header.Set("Accept", "application/vnd.schemaregistry.v1+json")

	if cfg.Avro.Username != "" {
		req.SetBasicAuth(cfg.Avro.Username, cfg.Avro.Password)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		rep.Warnings = append(rep.Warnings, report.Change{
			ID:             "avro_registry_unavailable",
			RuleID:         "AVRO_REGISTRY_UNAVAILABLE",
			Severity:       report.ChangeSeverity(report.SeverityWarning),
			Description:    "Could not reach Schema Registry: " + err.Error(),
			Recommendation: func(s string) *string { return &s }("Check that schema_registry_url is correct and the registry is reachable from CI."),
		})
		rep.Summary.WarningCount = 1
		compliance.Audit(rep)
		return rep, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		rep.Warnings = append(rep.Warnings, report.Change{
			ID:          "avro_registry_unavailable",
			RuleID:      "AVRO_REGISTRY_UNAVAILABLE",
			Severity:    report.ChangeSeverity(report.SeverityWarning),
			Description: fmt.Sprintf("Schema Registry returned HTTP %d", resp.StatusCode),
		})
		rep.Summary.WarningCount = 1
		compliance.Audit(rep)
		return rep, nil
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry response: %w", err)
	}

	var compatResp schemaRegistryCompatResponse
	if err := json.Unmarshal(respBody, &compatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal registry response: %w", err)
	}

	if compatResp.IsCompatible {
		compliance.Audit(rep)
		return rep, nil
	}

	desc := "Schema Registry reported incompatibility"
	if len(compatResp.Messages) > 0 {
		desc = strings.Join(compatResp.Messages, "; ")
	}

	rep.BreakingChanges = append(rep.BreakingChanges, report.Change{
		ID:             "avro_incompatible",
		RuleID:         "AVRO_INCOMPATIBLE",
		Severity:       report.ChangeSeverity(report.SeverityBreaking),
		Path:           subject,
		Description:    desc,
		Recommendation: func(s string) *string { return &s }("Review Avro schema evolution rules. Adding a field without a default value is backward-incompatible."),
	})
	rep.Summary.BreakingCount = 1
	compliance.Audit(rep)
	return rep, nil
}
