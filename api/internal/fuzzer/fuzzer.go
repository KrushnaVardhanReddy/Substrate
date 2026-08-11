// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package fuzzer

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
	"github.com/getkin/kin-openapi/openapi3"
)

// VulnerabilityReport represents a single vulnerability found by the fuzzer.
type VulnerabilityReport struct {
	Method   string
	Path     string
	Payload  map[string]interface{}
	Issue    string
	Severity string // e.g., "High", "Medium", "Low"
}

// Fuzzer is responsible for generating negative and boundary test payloads.
type Fuzzer struct {
	SpecPath string
	Queries  *sqlcgen.Queries
}

func NewFuzzer(specPath string, queries *sqlcgen.Queries) *Fuzzer {
	return &Fuzzer{SpecPath: specPath, Queries: queries}
}

// RunBackgroundFuzzing runs fuzzing in a goroutine and reports gaps.
func (f *Fuzzer) RunBackgroundFuzzing(ctx context.Context, results chan<- VulnerabilityReport) {
	go func() {
		defer close(results)
		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromFile(f.SpecPath)
		if err != nil {
			log.Printf("Fuzzer error: failed to load spec %s: %v", f.SpecPath, err)
			return
		}

		if err := doc.Validate(ctx); err != nil {
			log.Printf("Fuzzer error: invalid spec %s: %v", f.SpecPath, err)
			return
		}

		for path, pathItem := range doc.Paths.Map() {
			operations := map[string]*openapi3.Operation{
				"GET":    pathItem.Get,
				"POST":   pathItem.Post,
				"PUT":    pathItem.Put,
				"DELETE": pathItem.Delete,
				"PATCH":  pathItem.Patch,
			}

			for method, op := range operations {
				if op == nil {
					continue
				}

				if op.RequestBody != nil && op.RequestBody.Value != nil {
					content := op.RequestBody.Value.Content["application/json"]
					if content != nil && content.Schema != nil && content.Schema.Value != nil {
						schema := content.Schema.Value
						f.generateAndReportPayloads(ctx, method, path, schema, results)
					}
				}
			}
		}
	}()
}

func (f *Fuzzer) generateAndReportPayloads(ctx context.Context, method, path string, schema *openapi3.Schema, results chan<- VulnerabilityReport) {
	// SQLi payload
	sqliPayload := f.generateBasePayload(schema)
	for propName := range schema.Properties {
		sqliPayload[propName] = "' OR 1=1 --"
	}
	f.reportVulnerability(ctx, method, path, sqliPayload, "Schema Validation Gap: Potential SQL Injection vulnerability detected", "High", results)

	// Path Traversal payload
	ptPayload := f.generateBasePayload(schema)
	for propName := range schema.Properties {
		ptPayload[propName] = "../../../etc/passwd"
	}
	f.reportVulnerability(ctx, method, path, ptPayload, "Schema Validation Gap: Potential Path Traversal vulnerability detected", "High", results)

	// Null byte payload
	nullPayload := f.generateBasePayload(schema)
	for propName := range schema.Properties {
		nullPayload[propName] = "test\x00"
	}
	f.reportVulnerability(ctx, method, path, nullPayload, "Schema Validation Gap: Potential Null Byte Injection vulnerability detected", "Medium", results)

	// Extremely long string
	longStrPayload := f.generateBasePayload(schema)
	for propName, propRef := range schema.Properties {
		if propRef.Value != nil && propRef.Value.Type != nil && len(propRef.Value.Type.Slice()) > 0 && propRef.Value.Type.Slice()[0] == "string" {
			longStrPayload[propName] = strings.Repeat("A", 10000)
		}
	}
	f.reportVulnerability(ctx, method, path, longStrPayload, "Schema Validation Gap: Missing bounds check (Extremely Long String)", "Medium", results)
}

func (f *Fuzzer) reportVulnerability(ctx context.Context, method, path string, payload map[string]interface{}, issue, severity string, results chan<- VulnerabilityReport) {
	report := VulnerabilityReport{
		Method:   method,
		Path:     path,
		Payload:  payload,
		Issue:    issue,
		Severity: severity,
	}

	if f.Queries != nil {
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Fuzzer error: failed to marshal payload: %v", err)
		} else {
			err = f.Queries.InsertSchemaValidationGap(ctx, sqlcgen.InsertSchemaValidationGapParams{
				Method:   method,
				Path:     path,
				Payload:  payloadBytes,
				Issue:    issue,
				Severity: severity,
			})
			if err != nil {
				log.Printf("Fuzzer error: failed to insert schema validation gap: %v", err)
			}
		}
	}

	results <- report
}

func (f *Fuzzer) generateBasePayload(schema *openapi3.Schema) map[string]interface{} {
	payload := make(map[string]interface{})
	for propName, propRef := range schema.Properties {
		if propRef.Value == nil {
			continue
		}
		prop := propRef.Value

		if prop.Type != nil && len(prop.Type.Slice()) > 0 {
			switch prop.Type.Slice()[0] {
			case "string":
				payload[propName] = "test"
			case "integer", "number":
				minVal := 0.0
				if prop.Min != nil {
					minVal = *prop.Min
				}
				payload[propName] = minVal
			case "boolean":
				payload[propName] = true
			}
		}
	}
	return payload
}
