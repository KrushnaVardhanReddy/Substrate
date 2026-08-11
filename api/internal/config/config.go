// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type LLMConfig struct {
	Provider string
	APIKey   string
	BaseURL  string
	Model    string
}

func LoadLLMConfig() LLMConfig {
	provider := os.Getenv("LLM_PROVIDER")
	if provider == "" {
		provider = os.Getenv("SUBSTRATE_AI_PROVIDER")
	}
	if provider == "" {
		provider = "openai"
	}

	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("SUBSTRATE_AI_API_KEY")
	}
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}

	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL == "" {
		baseURL = os.Getenv("SUBSTRATE_AI_BASE_URL")
	}

	model := os.Getenv("LLM_MODEL")
	if model == "" {
		model = os.Getenv("SUBSTRATE_AI_MODEL")
	}
	if model == "" {
		model = "gpt-4o"
	}

	return LLMConfig{
		Provider: provider,
		APIKey:   apiKey,
		BaseURL:  baseURL,
		Model:    model,
	}
}

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
