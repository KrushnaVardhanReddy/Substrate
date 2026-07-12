package discovery

import (
	"context"
	"testing"
)

func TestOpenAPIScanner(t *testing.T) {
	tests := []struct {
		name        string
		files       map[string][]byte
		urlResolver func(string) string
		expected    []DependencyEdge
	}{
		{
			name: "scan openapi-generator-config.yaml",
			files: map[string][]byte{
				"openapi-generator-config.yaml": []byte(`
generatorName: typescript-axios
inputSpec: "https://raw.githubusercontent.com/myorg/backend-api/main/openapi.yaml"
`),
			},
			urlResolver: func(url string) string {
				if url == "https://raw.githubusercontent.com/myorg/backend-api/main/openapi.yaml" {
					return "myorg/backend-api"
				}
				return ""
			},
			expected: []DependencyEdge{
				{
					TargetRepo: "myorg/backend-api",
					Confidence: 50,
				},
			},
		},
		{
			name: "scan openapitools.json",
			files: map[string][]byte{
				"openapitools.json": []byte(`{
					"generator-cli": {
						"generators": {
							"users-client": {
								"inputSpec": "https://raw.githubusercontent.com/myorg/users-service/main/openapi.yaml"
							}
						}
					}
				}`),
			},
			urlResolver: func(url string) string {
				if url == "https://raw.githubusercontent.com/myorg/users-service/main/openapi.yaml" {
					return "myorg/users-service"
				}
				return ""
			},
			expected: []DependencyEdge{
				{
					TargetRepo: "myorg/users-service",
					Confidence: 50,
				},
			},
		},
		{
			name: "scan package.json openapi-generator",
			files: map[string][]byte{
				"package.json": []byte(`{
					"scripts": {
						"generate": "openapi-generator generate -i https://raw.githubusercontent.com/myorg/api/main/spec.yaml"
					}
				}`),
			},
			urlResolver: func(url string) string {
				if url == "https://raw.githubusercontent.com/myorg/api/main/spec.yaml" {
					return "myorg/api"
				}
				return ""
			},
			expected: []DependencyEdge{
				{
					TargetRepo: "myorg/api",
					Confidence: 50,
				},
			},
		},
		{
			name: "scan package.json swagger-codegen",
			files: map[string][]byte{
				"package.json": []byte(`{
					"scripts": {
						"generate": "swagger-codegen generate -i https://raw.githubusercontent.com/myorg/swagger/main/spec.yaml"
					}
				}`),
			},
			urlResolver: func(url string) string {
				if url == "https://raw.githubusercontent.com/myorg/swagger/main/spec.yaml" {
					return "myorg/swagger"
				}
				return ""
			},
			expected: []DependencyEdge{
				{
					TargetRepo: "myorg/swagger",
					Confidence: 50,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner := NewOpenAPIScanner()
			edges, err := scanner.Scan(context.Background(), "myorg/frontend", tt.files, tt.urlResolver)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(edges) != len(tt.expected) {
				t.Fatalf("expected %d edges, got %d", len(tt.expected), len(edges))
			}

			for i, expectedEdge := range tt.expected {
				if edges[i].TargetRepo != expectedEdge.TargetRepo {
					t.Errorf("expected TargetRepo %s, got %s", expectedEdge.TargetRepo, edges[i].TargetRepo)
				}
				if edges[i].Confidence != expectedEdge.Confidence {
					t.Errorf("expected Confidence %d, got %d", expectedEdge.Confidence, edges[i].Confidence)
				}
			}
		})
	}
}
