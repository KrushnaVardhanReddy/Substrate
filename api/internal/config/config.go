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

type SubstrateConfig struct {
	Discovery *Discovery `yaml:"discovery,omitempty"`
	Metadata  *Metadata  `yaml:"metadata,omitempty"`
}

func Parse(content []byte) (*SubstrateConfig, error) {
	var cfg SubstrateConfig
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}
	return &cfg, nil
}
