// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package fuzzer

import (
	"context"
	"os"
	"testing"
)

func TestRunBackgroundFuzzing(t *testing.T) {
	// Create a dummy OpenAPI spec
	specContent := `{
		"openapi": "3.0.0",
		"info": {
			"title": "Test API",
			"version": "1.0.0"
		},
		"paths": {
			"/test": {
				"post": {
					"requestBody": {
						"content": {
							"application/json": {
								"schema": {
									"type": "object",
									"properties": {
										"name": {
											"type": "string"
										},
										"id": {
											"type": "integer"
										}
									}
								}
							}
						}
					},
					"responses": {
						"200": {
							"description": "OK"
						}
					}
				}
			}
		}
	}`

	tmpFile, err := os.CreateTemp("", "openapi-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(specContent)); err != nil {
		t.Fatalf("Failed to write temp file: %v", err)
	}
	tmpFile.Close()

	f := NewFuzzer(tmpFile.Name(), nil)
	results := make(chan VulnerabilityReport, 10)

	f.RunBackgroundFuzzing(context.Background(), results)

	var reports []VulnerabilityReport
	for report := range results {
		reports = append(reports, report)
	}

	if len(reports) == 0 {
		t.Fatalf("Expected vulnerability reports, got 0")
	}

	foundSQLi := false
	foundPathTraversal := false
	for _, r := range reports {
		if r.Method != "POST" || r.Path != "/test" {
			t.Errorf("Unexpected method/path: %s %s", r.Method, r.Path)
		}
		if r.Issue == "Schema Validation Gap: Potential SQL Injection vulnerability detected" {
			foundSQLi = true
			if r.Payload["name"] != "' OR 1=1 --" {
				t.Errorf("Expected SQLi payload, got %v", r.Payload["name"])
			}
		}
		if r.Issue == "Schema Validation Gap: Potential Path Traversal vulnerability detected" {
			foundPathTraversal = true
		}
	}

	if !foundSQLi {
		t.Errorf("Did not find expected SQLi vulnerability report")
	}
	if !foundPathTraversal {
		t.Errorf("Did not find expected Path Traversal vulnerability report")
	}
}
