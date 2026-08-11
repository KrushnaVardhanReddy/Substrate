// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package discovery

import (
	"encoding/json"
	"regexp"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"gopkg.in/yaml.v3"
)

type EventScanner struct{}

func NewEventScanner() *EventScanner {
	return &EventScanner{}
}

// ScanAsyncAPI parses an AsyncAPI spec and extracts cross-repo dependencies via $ref.
// According to docs/specs/phase-5/dependency-discovery.md, this contributes +35 pts confidence.
func (s *EventScanner) ScanAsyncAPI(content []byte, sourceRepo string) []Edge {
	var spec map[string]interface{}

	// Try YAML first
	err := yaml.Unmarshal(content, &spec)
	if err != nil {
		// Fallback to JSON
		err = json.Unmarshal(content, &spec)
		if err != nil {
			return nil // Could not parse
		}
	}

	var edges []Edge

	// Very simple recursive parser to find $refs
	var findRefs func(node interface{})
	findRefs = func(node interface{}) {
		switch v := node.(type) {
		case map[string]interface{}:
			for key, val := range v {
				if key == "$ref" {
					if refStr, ok := val.(string); ok {
						// Extract github repo from URL: https://github.com/myorg/orders-service/...
						re := regexp.MustCompile(`https://github\.com/([^/]+/[^/]+)/`)
						matches := re.FindStringSubmatch(refStr)
						if len(matches) > 1 {
							targetRepo := matches[1]
							if targetRepo != sourceRepo {
								edges = append(edges, Edge{
									SourceRepo: sourceRepo,
									TargetRepo: targetRepo,
									Confidence: 35,
									Signal:     "asyncapi_$ref",
								})
							}
						}
					}
				} else {
					findRefs(val)
				}
			}
		case []interface{}:
			for _, item := range v {
				findRefs(item)
			}
		}
	}

	findRefs(spec)
	return edges
}

// ScanTerraformKafka parses Terraform HCL looking for confluent_kafka_topic blocks.
// According to docs/specs/phase-5/dependency-discovery.md, this contributes +20 pts confidence.
func (s *EventScanner) ScanTerraformKafka(content []byte, sourceRepo string) []Edge {
	file, diags := hclsyntax.ParseConfig(content, "main.tf", hcl.Pos{Line: 1, Column: 1})
	if diags.HasErrors() {
		return nil
	}

	var edges []Edge
	if file.Body != nil {
		body := file.Body.(*hclsyntax.Body)
		for _, block := range body.Blocks {
			if block.Type == "resource" && len(block.Labels) >= 2 && block.Labels[0] == "confluent_kafka_topic" {
				if block.Body != nil {
					if attr, exists := block.Body.Attributes["topic_name"]; exists {
						val, diags := attr.Expr.Value(nil)
						if !diags.HasErrors() && val.Type().IsPrimitiveType() {
							edges = append(edges, Edge{
								SourceRepo: sourceRepo,
								TargetRepo: sourceRepo,
								Confidence: 20,
								Signal:     "kafka_topic:" + val.AsString(),
							})
						}
					}
				}
			}
		}
	}

	return edges
}
