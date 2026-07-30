package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Discovery struct {
	MatchPatterns []string `yaml:"match_patterns"`
}

type Metadata struct {
	Type         string   `yaml:"type,omitempty" json:"type,omitempty"`
	Team         string   `yaml:"team,omitempty" json:"team,omitempty"`
	Databases    []string `yaml:"databases,omitempty" json:"databases,omitempty"`
	Owner        string   `yaml:"owner,omitempty" json:"owner,omitempty"`
	SlackChannel string   `yaml:"slack_channel,omitempty" json:"slack_channel,omitempty"`
	PagerDuty    string   `yaml:"pagerduty,omitempty" json:"pagerduty,omitempty"`
	PM           string   `yaml:"pm,omitempty" json:"pm,omitempty"`
	SLATier      string   `yaml:"sla_tier,omitempty" json:"sla_tier,omitempty"`
	NodeColor    string   `yaml:"node_color,omitempty" json:"node_color,omitempty"`
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

type Gateway struct {
	Type       string `yaml:"type"`
	InfraRepo  string `yaml:"infra_repo"`
	OutputPath string `yaml:"output_path"`
}

type SubstrateConfig struct {
	Discovery  *Discovery       `yaml:"discovery,omitempty"`
	Metadata   *Metadata        `yaml:"metadata,omitempty"`
	SchemaType string           `yaml:"schema_type,omitempty"`
	BaseSchema string           `yaml:"base_schema,omitempty"`
	HeadSchema string           `yaml:"head_schema,omitempty"`
	Consumers  []ConsumerConfig `yaml:"consumers,omitempty"`
	Gateway    *Gateway         `yaml:"gateway,omitempty"`
}

func Parse(content []byte) (*SubstrateConfig, error) {
	var cfg SubstrateConfig
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}
	return &cfg, nil
}
