package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Discovery struct {
	MatchPatterns []string `yaml:"match_patterns"`
}

type Metadata struct {
	Type      string   `yaml:"type,omitempty"`
	Team      string   `yaml:"team,omitempty"`
	Databases []string `yaml:"databases,omitempty"`
}

type OverrideConfig struct {
	RuleID string `yaml:"rule_id"`
}

type ConsumerConfig struct {
	Name             string           `yaml:"name"`
	ProviderRepo     string           `yaml:"provider_repo"`
	SchemaType       string           `yaml:"schema_type"`
	ProviderSpecPath string           `yaml:"provider_spec_path"`
	ProviderBranch   string           `yaml:"provider_branch"`
	Overrides        []OverrideConfig `yaml:"overrides,omitempty"`
}

type SubstrateConfig struct {
	Discovery  *Discovery       `yaml:"discovery,omitempty"`
	Metadata   *Metadata        `yaml:"metadata,omitempty"`
	SchemaType string           `yaml:"schema_type,omitempty"`
	BaseSchema string           `yaml:"base_schema,omitempty"`
	HeadSchema string           `yaml:"head_schema,omitempty"`
	Consumers  []ConsumerConfig `yaml:"consumers,omitempty"`
}

func Parse(content []byte) (*SubstrateConfig, error) {
	var cfg SubstrateConfig
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}
	return &cfg, nil
}
