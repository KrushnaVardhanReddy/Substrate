package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

	repoRoot := filepath.Dir(path)
	absSpecPath := filepath.Join(repoRoot, config.SpecPath)
	if _, err := os.Stat(absSpecPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("substrate.yaml: spec_path '%s' not found", config.SpecPath)
	}

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
