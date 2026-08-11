// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package compliance

import (
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/config"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/stretchr/testify/assert"
)

func TestScanOpenAPISchema(t *testing.T) {
	tests := []struct {
		name         string
		schema       *openapi3.Schema
		cfg          *config.SubstrateConfig
		expectedTags map[string][]string
	}{
		{
			name: "Detect GDPR from property name",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"object"},
				Properties: openapi3.Schemas{
					"ssn": &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"string"},
						},
					},
				},
			},
			cfg: nil,
			expectedTags: map[string][]string{
				"ssn": {"GDPR"},
			},
		},
		{
			name: "Detect HIPAA from description",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"object"},
				Properties: openapi3.Schemas{
					"history": &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type:        &openapi3.Types{"string"},
							Description: "Patient medical history",
						},
					},
				},
			},
			cfg: nil,
			expectedTags: map[string][]string{
				"history": {"HIPAA"},
			},
		},
		{
			name: "Custom pattern from config",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"object"},
				Properties: openapi3.Schemas{
					"bank_routing": &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"string"},
						},
					},
				},
			},
			cfg: &config.SubstrateConfig{
				Compliance: &config.ComplianceConfig{
					Patterns: map[string]string{
						"FINANCE": `(?i)(bank.?routing|account.?number)`,
					},
				},
			},
			expectedTags: map[string][]string{
				"bank_routing": {"FINANCE"},
			},
		},
		{
			name: "Override default pattern from config",
			schema: &openapi3.Schema{
				Type: &openapi3.Types{"object"},
				Properties: openapi3.Schemas{
					"email": &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"string"},
						},
					},
					"custom_ssn": &openapi3.SchemaRef{
						Value: &openapi3.Schema{
							Type: &openapi3.Types{"string"},
						},
					},
				},
			},
			cfg: &config.SubstrateConfig{
				Compliance: &config.ComplianceConfig{
					Patterns: map[string]string{
						"GDPR": `(?i)(custom_ssn)`,
					},
				},
			},
			expectedTags: map[string][]string{
				"custom_ssn": {"GDPR"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := &openapi3.T{
				Components: &openapi3.Components{
					Schemas: map[string]*openapi3.SchemaRef{
						"TestSchema": {
							Value: tt.schema,
						},
					},
				},
			}

			ScanOpenAPISchema(doc, tt.cfg)

			for prop, tags := range tt.expectedTags {
				propSchema := doc.Components.Schemas["TestSchema"].Value.Properties[prop].Value
				assert.NotNil(t, propSchema.Extensions)

				// order might not be guaranteed, but we usually only have 1 match in tests
				assert.ElementsMatch(t, tags, propSchema.Extensions["x-compliance"])
			}

			// check that properties not in expected tags have no extensions
			for prop, propRef := range doc.Components.Schemas["TestSchema"].Value.Properties {
				if _, ok := tt.expectedTags[prop]; !ok {
					assert.Nil(t, propRef.Value.Extensions["x-compliance"])
				}
			}
		})
	}
}
