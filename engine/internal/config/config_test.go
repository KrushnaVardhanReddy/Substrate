package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	tempDir := t.TempDir()

	// Create a dummy spec file for tests
	specPath := filepath.Join(tempDir, "openapi.yaml")
	if err := os.WriteFile(specPath, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
		errMsg      string
	}{
		{
			name: "Valid Config",
			yamlContent: `
service: my-service
spec_path: openapi.yaml
overrides:
  - rule_id: FIELD_REMOVED
    path: components.schemas.Customer
    reason: "This is a valid reason that is long enough."
    approved_by: dev@company.com
    expires: 2099-01-01
`,
			wantErr: false,
		},
		{
			name: "Missing Service",
			yamlContent: `
spec_path: openapi.yaml
`,
			wantErr: true,
			errMsg:  "substrate.yaml: 'service' is required",
		},
		{
			name: "Missing Spec Path",
			yamlContent: `
service: my-service
`,
			wantErr: true,
			errMsg:  "substrate.yaml: 'spec_path' is required",
		},
		{
			name: "Missing Spec File",
			yamlContent: `
service: my-service
spec_path: missing.yaml
`,
			wantErr: true,
			errMsg:  "substrate.yaml: spec_path 'missing.yaml' not found",
		},
		{
			name: "Unknown Rule ID",
			yamlContent: `
service: my-service
spec_path: openapi.yaml
overrides:
  - rule_id: MY_CUSTOM_RULE
    path: components.schemas.Customer
    reason: "This is a valid reason that is long enough."
    approved_by: dev@company.com
    expires: 2099-01-01
`,
			wantErr: true,
			errMsg:  "substrate.yaml: unknown rule_id 'MY_CUSTOM_RULE'",
		},
		{
			name: "Reason Too Short",
			yamlContent: `
service: my-service
spec_path: openapi.yaml
overrides:
  - rule_id: FIELD_REMOVED
    path: components.schemas.Customer
    reason: "Too short"
    approved_by: dev@company.com
    expires: 2099-01-01
`,
			wantErr: true,
			errMsg:  "substrate.yaml: override reason too short (min 20 chars)",
		},
		{
			name: "Expired More Than 30 Days",
			yamlContent: `
service: my-service
spec_path: openapi.yaml
overrides:
  - rule_id: FIELD_REMOVED
    path: components.schemas.Customer
    reason: "This is a valid reason that is long enough."
    approved_by: dev@company.com
    expires: 2020-01-01
`,
			wantErr: true,
			errMsg:  "substrate.yaml: override for 'FIELD_REMOVED' expired on 2020-01-01 — remove or renew",
		},
		{
			name: "Unknown Schema Type",
			yamlContent: `
service: my-service
spec_path: openapi.yaml
schema_type: avro
`,
			wantErr: true,
			errMsg:  "substrate.yaml: unknown schema_type 'avro'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tempDir, "substrate.yaml")
			if err := os.WriteFile(configPath, []byte(tt.yamlContent), 0644); err != nil {
				t.Fatal(err)
			}

			_, err := LoadConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg && !contains(err.Error(), tt.errMsg) {
					t.Errorf("LoadConfig() error msg = %v, want %v", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	// A basic string contains function just for the test msg
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}

func TestIsOverrideActive(t *testing.T) {
	futureDate := time.Now().Add(24 * time.Hour).Format("2006-01-02")
	pastDate := time.Now().Add(-24 * time.Hour).Format("2006-01-02")

	config := &SubstrateConfig{
		Overrides: []Override{
			{
				RuleID:  "FIELD_REMOVED",
				Path:    "components.schemas.Customer",
				Expires: futureDate,
			},
			{
				RuleID:  "ENUM_VALUE_ADDED",
				Path:    "components.schemas.OrderStatus",
				Expires: pastDate,
			},
		},
	}

	tests := []struct {
		name   string
		ruleID string
		path   string
		want   bool
	}{
		{
			name:   "Active override (future expiry)",
			ruleID: "FIELD_REMOVED",
			path:   "components.schemas.Customer",
			want:   true,
		},
		{
			name:   "Expired override (past expiry)",
			ruleID: "ENUM_VALUE_ADDED",
			path:   "components.schemas.OrderStatus",
			want:   false,
		},
		{
			name:   "No matching rule/path",
			ruleID: "FIELD_REMOVED",
			path:   "components.schemas.Other",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := config.IsOverrideActive(tt.ruleID, tt.path); got != tt.want {
				t.Errorf("IsOverrideActive() = %v, want %v", got, tt.want)
			}
		})
	}
}
