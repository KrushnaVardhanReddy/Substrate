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
schema_type: unknownformat
`,
			wantErr: true,
			errMsg:  "substrate.yaml: unknown schema_type 'unknownformat' (supported: openapi, sql, graphql, protobuf, asyncapi, avro)",
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

func TestConsumerParsing(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")
	if err := os.WriteFile(specPath, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
		errMsg      string
		validate    func(t *testing.T, c *SubstrateConfig)
	}{
		{
			name: "no consumers block",
			yamlContent: `
service: my-api
schema_type: openapi
spec_path: openapi.yaml
`,
			validate: func(t *testing.T, c *SubstrateConfig) {
				if len(c.Consumers) != 0 {
					t.Errorf("Expected 0 consumers, got %d", len(c.Consumers))
				}
				if c.HasConsumers() {
					t.Errorf("Expected HasConsumers() to be false")
				}
			},
		},
		{
			name: "single consumer",
			yamlContent: `
service: frontend
spec_path: openapi.yaml
consumers:
  - name: users-api
    provider_repo: myorg/backend-api
    schema_type: openapi
    provider_spec_path: api/openapi.yaml
    provider_branch: main
`,
			validate: func(t *testing.T, c *SubstrateConfig) {
				if len(c.Consumers) != 1 {
					t.Fatalf("Expected 1 consumer, got %d", len(c.Consumers))
				}
				if c.Consumers[0].Name != "users-api" {
					t.Errorf("Expected consumer name 'users-api', got '%s'", c.Consumers[0].Name)
				}
				if c.Consumers[0].ProviderRepo != "myorg/backend-api" {
					t.Errorf("Expected provider repo 'myorg/backend-api', got '%s'", c.Consumers[0].ProviderRepo)
				}
				if c.Consumers[0].DefaultedBranch() != "main" {
					t.Errorf("Expected defaulted branch 'main', got '%s'", c.Consumers[0].DefaultedBranch())
				}
				if !c.HasConsumers() {
					t.Errorf("Expected HasConsumers() to be true")
				}
			},
		},
		{
			name: "consumer with missing branch defaults to main",
			yamlContent: `
service: frontend
spec_path: openapi.yaml
consumers:
  - name: users-api
    provider_repo: myorg/backend-api
    schema_type: openapi
    provider_spec_path: api/openapi.yaml
`,
			validate: func(t *testing.T, c *SubstrateConfig) {
				if len(c.Consumers) != 1 {
					t.Fatalf("Expected 1 consumer, got %d", len(c.Consumers))
				}
				if c.Consumers[0].DefaultedBranch() != "main" {
					t.Errorf("Expected defaulted branch 'main', got '%s'", c.Consumers[0].DefaultedBranch())
				}
			},
		},
		{
			name: "multiple consumers",
			yamlContent: `
service: frontend
spec_path: openapi.yaml
consumers:
  - name: users-api
    provider_repo: myorg/backend-api
    schema_type: openapi
    provider_spec_path: api/openapi.yaml
  - name: payments-api
    provider_repo: myorg/payments-service
    schema_type: openapi
    provider_spec_path: openapi.yaml
`,
			validate: func(t *testing.T, c *SubstrateConfig) {
				if len(c.Consumers) != 2 {
					t.Fatalf("Expected 2 consumers, got %d", len(c.Consumers))
				}
				if c.Consumers[0].Name != "users-api" {
					t.Errorf("Expected 1st consumer 'users-api', got '%s'", c.Consumers[0].Name)
				}
				if c.Consumers[1].Name != "payments-api" {
					t.Errorf("Expected 2nd consumer 'payments-api', got '%s'", c.Consumers[1].Name)
				}
			},
		},
		{
			name: "existing substrate.yaml fields still parse correctly",
			yamlContent: `
service: my-service
spec_path: openapi.yaml
schema_type: openapi
overrides:
  - rule_id: FIELD_REMOVED
    path: components.schemas.Customer
    reason: "This is a valid reason that is long enough."
    approved_by: dev@company.com
    expires: 2099-01-01
consumers:
  - name: "users-api"
    provider_repo: "myorg/backend-api"
    schema_type: openapi
    provider_spec_path: "api/openapi.yaml"
    provider_branch: "main"
`,
			validate: func(t *testing.T, c *SubstrateConfig) {
				if c.Service != "my-service" {
					t.Errorf("Expected service 'my-service', got '%s'", c.Service)
				}
				if c.SchemaType != "openapi" {
					t.Errorf("Expected schema_type 'openapi', got '%s'", c.SchemaType)
				}
				if len(c.Overrides) != 1 {
					t.Fatalf("Expected 1 override, got %d", len(c.Overrides))
				}
				if c.Overrides[0].RuleID != "FIELD_REMOVED" {
					t.Errorf("Expected rule_id 'FIELD_REMOVED', got '%s'", c.Overrides[0].RuleID)
				}
				if len(c.Consumers) != 1 {
					t.Fatalf("Expected 1 consumer, got %d", len(c.Consumers))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tempDir, "substrate.yaml")
			if err := os.WriteFile(configPath, []byte(tt.yamlContent), 0644); err != nil {
				t.Fatal(err)
			}

			config, err := LoadConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg && !contains(err.Error(), tt.errMsg) {
					t.Errorf("LoadConfig() error msg = %v, want %v", err.Error(), tt.errMsg)
				}
			}

			if err == nil && tt.validate != nil {
				tt.validate(t, config)
			}
		})
	}
}

func TestAvroParsing(t *testing.T) {
	tempDir := t.TempDir()
	specPath := filepath.Join(tempDir, "openapi.yaml")
	if err := os.WriteFile(specPath, []byte("dummy"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		yamlContent string
		wantErr     bool
		errMsg      string
		validate    func(t *testing.T, c *SubstrateConfig)
	}{
		{
			name: "avro config parsed",
			yamlContent: `
service: payments
schema_type: avro
spec_path: schemas/payment.avsc
avro:
  schema_registry_url: https://registry.example.com
  subject: payments-value
  username: user
  password: pass
`,
			validate: func(t *testing.T, c *SubstrateConfig) {
				if c.Avro == nil {
					t.Fatal("Expected cfg.Avro != nil")
				}
				if c.Avro.SchemaRegistryURL != "https://registry.example.com" {
					t.Errorf("Expected URL 'https://registry.example.com', got '%s'", c.Avro.SchemaRegistryURL)
				}
				if c.Avro.Subject != "payments-value" {
					t.Errorf("Expected subject 'payments-value', got '%s'", c.Avro.Subject)
				}
				if c.Avro.Username != "user" {
					t.Errorf("Expected username 'user', got '%s'", c.Avro.Username)
				}
			},
		},
		{
			name: "avro config omitted",
			yamlContent: `
service: my-api
schema_type: openapi
spec_path: openapi.yaml
`,
			validate: func(t *testing.T, c *SubstrateConfig) {
				if c.Avro != nil {
					t.Errorf("Expected cfg.Avro == nil, got %v", c.Avro)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tempDir, "substrate.yaml")
			if err := os.WriteFile(configPath, []byte(tt.yamlContent), 0644); err != nil {
				t.Fatal(err)
			}

			// Needs dummy avsc for testing spec exists check
			if err := os.MkdirAll(filepath.Join(tempDir, "schemas"), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(tempDir, "schemas/payment.avsc"), []byte("dummy"), 0644); err != nil {
				t.Fatal(err)
			}

			config, err := LoadConfig(configPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg && !contains(err.Error(), tt.errMsg) {
					t.Errorf("LoadConfig() error msg = %v, want %v", err.Error(), tt.errMsg)
				}
			}

			if err == nil && tt.validate != nil {
				tt.validate(t, config)
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
