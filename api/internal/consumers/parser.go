// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package consumers

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type ConsumerPath struct {
	Path   string   `yaml:"path"`
	Fields []string `yaml:"fields"`
}

type Manifest struct {
	SchemaVersion string         `yaml:"schema_version"`
	Provider      string         `yaml:"provider"`
	Consumes      []ConsumerPath `yaml:"consumes"`
}

func Parse(data []byte) (*Manifest, error) {
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("failed to parse yaml: %w", err)
	}

	if m.SchemaVersion == "" {
		return nil, fmt.Errorf("missing schema_version")
	}

	if m.Provider == "" {
		return nil, fmt.Errorf("missing provider")
	}

	if len(m.Consumes) == 0 {
		return nil, fmt.Errorf("missing consumes")
	}

	for _, c := range m.Consumes {
		if c.Path == "" {
			return nil, fmt.Errorf("missing path in consumes")
		}
		if len(c.Fields) == 0 {
			return nil, fmt.Errorf("missing fields in consumes")
		}
	}

	return &m, nil
}
