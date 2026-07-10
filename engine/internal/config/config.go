package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type ConsumerDependency struct {
	Name             string `yaml:"name"`
	ProviderRepo     string `yaml:"provider_repo"`
	SchemaType       string `yaml:"schema_type"`
	ProviderSpecPath string `yaml:"provider_spec_path"`
	ProviderBranch   string `yaml:"provider_branch"`
}

func (d *ConsumerDependency) DefaultedBranch() string {
	if d.ProviderBranch == "" {
		return "main"
	}
	return d.ProviderBranch
}

type Owner struct {
	Team    string `yaml:"team"`
	Contact string `yaml:"contact,omitempty"`
}

type Override struct {
	RuleID     string `yaml:"rule_id"`
	Path       string `yaml:"path"`
	Reason     string `yaml:"reason"`
	ApprovedBy string `yaml:"approved_by"`
	Expires    string `yaml:"expires"`
}

type AvroConfig struct {
	SchemaRegistryURL string `yaml:"schema_registry_url"`
	Subject           string `yaml:"subject,omitempty"`
	Username          string `yaml:"username,omitempty"`
	Password          string `yaml:"password,omitempty"`
}

type SubstrateConfig struct {
	Version    string               `yaml:"version"`
	Service    string               `yaml:"service"`
	SchemaType string               `yaml:"schema_type,omitempty"`
	Mode       string               `yaml:"mode,omitempty"`
	SpecPath   string               `yaml:"spec_path"`
	Owners     []Owner              `yaml:"owners,omitempty"`
	Overrides  []Override           `yaml:"overrides,omitempty"`
	Consumers  []ConsumerDependency `yaml:"consumers,omitempty"`
	Avro       *AvroConfig          `yaml:"avro,omitempty"`
}

func (c *SubstrateConfig) HasConsumers() bool {
	return len(c.Consumers) > 0
}

var KnownRules = map[string]bool{
	"FIELD_REMOVED":                     true,
	"FIELD_ADDED_OPTIONAL":              true,
	"FIELD_RENAMED":                     true,
	"FIELD_DEPRECATED":                  true,
	"REQUIRED_FIELD_ADDED":              true,
	"REQUIRED_FIELD_MADE_OPTIONAL":      true,
	"ENUM_VALUE_REMOVED":                true,
	"ENUM_VALUE_ADDED":                  true,
	"ENUM_TYPE_CHANGED":                 true,
	"ENDPOINT_REMOVED":                  true,
	"ENDPOINT_ADDED":                    true,
	"ENDPOINT_DEPRECATED":               true,
	"METHOD_REMOVED":                    true,
	"METHOD_ADDED":                      true,
	"REQUEST_BODY_REMOVED":              true,
	"REQUEST_BODY_MADE_REQUIRED":        true,
	"REQUEST_BODY_CONTENT_TYPE_REMOVED": true,
	"REQUEST_BODY_CONTENT_TYPE_ADDED":   true,
	"RESPONSE_CODE_REMOVED":             true,
	"RESPONSE_CODE_ADDED":               true,
	"RESPONSE_SCHEMA_FIELD_REMOVED":     true,
	"RESPONSE_SCHEMA_FIELD_ADDED":       true,
	"PARAMETER_REMOVED":                 true,
	"PARAMETER_ADDED_REQUIRED":          true,
	"PARAMETER_ADDED_OPTIONAL":          true,
	"PARAMETER_MADE_REQUIRED":           true,
	"PARAMETER_TYPE_CHANGED":            true,
	"SECURITY_SCHEME_REMOVED":           true,
	"SECURITY_REQUIREMENT_ADDED":        true,
	"SECURITY_REQUIREMENT_REMOVED":      true,
	"FIELD_TYPE_CHANGED":                true,
	"FIELD_FORMAT_CHANGED":              true,
	"FIELD_NULLABLE_CHANGED":            true,
	"STABILITY_LEVEL_CHANGED":           true,
	"SPEC_INVALID_REF":                  true,
	"SPEC_INVALID_TYPE":                 true,
	"SPEC_MISSING_REQUIRED_FIELD":       true,
	// SQL Rules
	"TABLE_REMOVED":                     true,
	"TABLE_RENAMED":                     true,
	"TABLE_ADDED":                       true,
	"COLUMN_REMOVED":                    true,
	"COLUMN_RENAMED":                    true,
	"COLUMN_ADDED_NOT_NULL_NO_DEFAULT":  true,
	"COLUMN_ADDED_NULLABLE":             true,
	"COLUMN_TYPE_WIDENED":               true,
	"COLUMN_TYPE_CHANGED":               true,
	"COLUMN_MADE_NOT_NULL":              true,
	"COLUMN_NULLABLE_CHANGED":           true,
	"COLUMN_DEFAULT_REMOVED":            true,
	"COLUMN_DEFAULT_CHANGED":            true,
	"PRIMARY_KEY_CHANGED":               true,
	"CONSTRAINT_REMOVED":                true,
	"CONSTRAINT_ADDED_UNIQUE":           true,
	"CONSTRAINT_ADDED_CHECK":            true,
	"FOREIGN_KEY_ADDED":                 true,
	"INDEX_REMOVED":                     true,
	"INDEX_ADDED":                       true,
	"VIEW_REMOVED":                      true,
	"VIEW_ADDED":                        true,
	"VIEW_COLUMN_REMOVED":               true,
	"VIEW_DEFINITION_CHANGED":           true,
	"SQL_ENUM_TYPE_REMOVED":             true,
	"SQL_ENUM_VALUE_REMOVED":            true,
	"SQL_ENUM_VALUE_ADDED":              true,
	// GraphQL Rules
	"GQL_TYPE_REMOVED":                  true,
	"GQL_FIELD_REMOVED":                 true,
	"GQL_FIELD_TYPE_CHANGED":            true,
	"GQL_UNION_MEMBER_REMOVED":          true,
	"GQL_ENUM_VALUE_REMOVED":            true,
	"GQL_INTERFACE_REMOVED":             true,
	"GQL_ARGUMENT_REMOVED":              true,
	"GQL_ARGUMENT_TYPE_CHANGED":         true,
	"GQL_REQUIRED_ARGUMENT_ADDED":       true,
	"GQL_INPUT_FIELD_ADDED_REQUIRED":    true,
	"GQL_OPTIONAL_ARGUMENT_ADDED":       true,
	"GQL_INPUT_FIELD_ADDED_OPTIONAL":    true,
	"GQL_DIRECTIVE_REMOVED":             true,
	"GQL_DIRECTIVE_LOCATION_REMOVED":    true,
	"GQL_FIELD_DEPRECATED":              true,
	"GQL_ENUM_VALUE_DEPRECATED":         true,
}

func LoadConfig(path string) (*SubstrateConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, err // Let caller handle if they want defaults
		}
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var config SubstrateConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	// Default version
	if config.Version == "" {
		config.Version = "1"
	}

	// Default mode
	if config.Mode != "strict" && config.Mode != "legacy" {
		config.Mode = "strict"
	}

	// Validation
	if config.Service == "" {
		return nil, fmt.Errorf("substrate.yaml: 'service' is required")
	}
	if config.SpecPath == "" {
		return nil, fmt.Errorf("substrate.yaml: 'spec_path' is required")
	}

	// Removed os.Stat check for absSpecPath because the engine runs in a stateless HTTP context
	// where the spec file is not physically on disk next to the temporary config file.

	if config.SchemaType != "" {
		switch config.SchemaType {
		case "openapi", "sql", "graphql", "protobuf", "asyncapi", "avro":
			// valid
		default:
			return nil, fmt.Errorf("substrate.yaml: unknown schema_type '%s' (supported: openapi, sql, graphql, protobuf, asyncapi, avro)", config.SchemaType)
		}
	}

	for _, o := range config.Overrides {
		if !KnownRules[o.RuleID] {
			return nil, fmt.Errorf("substrate.yaml: unknown rule_id '%s'", o.RuleID)
		}

		if len(o.Reason) < 20 {
			return nil, fmt.Errorf("substrate.yaml: override reason too short (min 20 chars)")
		}

		t, err := time.Parse("2006-01-02", o.Expires)
		if err != nil {
			return nil, fmt.Errorf("substrate.yaml: invalid expires format '%s' for rule '%s'", o.Expires, o.RuleID)
		}

		nowStr := time.Now().UTC().Format("2006-01-02")
		now, _ := time.Parse("2006-01-02", nowStr)

		if now.Sub(t) > 30*24*time.Hour {
			return nil, fmt.Errorf("substrate.yaml: override for '%s' expired on %s — remove or renew", o.RuleID, o.Expires)
		}
	}

	return &config, nil
}

func (c *SubstrateConfig) IsOverrideActive(ruleID, path string) bool {
	for _, o := range c.Overrides {
		if o.RuleID == ruleID && o.Path == path {
			t, err := time.Parse("2006-01-02", o.Expires)
			if err == nil {
				nowStr := time.Now().UTC().Format("2006-01-02")
				now, _ := time.Parse("2006-01-02", nowStr)
				if !now.After(t) {
					return true
				}
			}
		}
	}
	return false
}
