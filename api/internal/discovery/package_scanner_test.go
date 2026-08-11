// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package discovery

import (
	"context"
	"testing"
)

func TestPackageScanner(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string][]byte
		urlResolver func(string) string
		expected    []DependencyEdge
	}{
		{
			name: "scan package.json",
			files: map[string][]byte{
				"package.json": []byte(`{
					"dependencies": {
						"@myorg/backend-sdk": "1.0.0",
						"react": "18.2.0"
					},
					"devDependencies": {
						"@myorg/test-utils": "2.0.0"
					}
				}`),
			},
			urlResolver: func(url string) string {
				if url == "@myorg/backend-sdk" {
					return "myorg/backend-api"
				}
				if url == "@myorg/test-utils" {
					return "myorg/test-utils"
				}
				return ""
			},
			expected: []DependencyEdge{
				{
					TargetRepo: "myorg/backend-api", // Order depends on map iteration, we'll sort or check presence below
					Confidence: 40,
				},
				{
					TargetRepo: "myorg/test-utils",
					Confidence: 40,
				},
			},
		},
		{
			name: "scan go.mod",
			files: map[string][]byte{
				"go.mod": []byte(`module github.com/myorg/frontend

go 1.20

require (
	github.com/myorg/backend-api v1.2.3
	github.com/stretchr/testify v1.8.0
)

require github.com/myorg/other-api v0.1.0
`),
			},
			urlResolver: func(url string) string {
				if url == "github.com/myorg/backend-api" {
					return "myorg/backend-api"
				}
				if url == "github.com/myorg/other-api" {
					return "myorg/other-api"
				}
				return ""
			},
			expected: []DependencyEdge{
				{
					TargetRepo: "myorg/backend-api",
					Confidence: 40,
				},
				{
					TargetRepo: "myorg/other-api",
					Confidence: 40,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewPackageScanner()
			edges, err := scanner.Scan(context.Background(), "myorg/frontend", tt.files, tt.urlResolver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(edges) != len(tt.expected) {
				t.Fatalf("expected %d edges, got %d", len(tt.expected), len(edges))
			}

			// Map target repos for easier assertion as array ordering might be non-deterministic for package.json
			foundRepos := make(map[string]bool)
			for _, edge := range edges {
				foundRepos[edge.TargetRepo] = true
				if edge.Confidence != 40 {
					t.Errorf("expected Confidence 40, got %d", edge.Confidence)
				}
			}

			for _, expectedEdge := range tt.expected {
				if !foundRepos[expectedEdge.TargetRepo] {
					t.Errorf("missing expected target repo: %s", expectedEdge.TargetRepo)
				}
			}
		})
	}
}
