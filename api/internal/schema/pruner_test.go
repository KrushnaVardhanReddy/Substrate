package schema

import (
	"context"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
)

type mockAIClient struct{}

func (m *mockAIClient) EmbedText(ctx context.Context, text string) ([]float64, error) {
	if text == "find user" || text == "/users Get users Get user info" {
		return []float64{1, 0, 0}, nil
	}
	if text == "/users/{id} Get user by id Get user by id" {
		return []float64{0.9, 0.1, 0}, nil
	}
	if text == "/login Post login User login" {
		return []float64{0, 1, 0}, nil
	}
	if text == "/posts Get posts Get posts" {
		return []float64{0, 0, 1}, nil
	}
	// Missing summary/desc
	if text == "/empty" {
		return []float64{0, 0.1, 0}, nil
	}

	return []float64{0.0, 0.1, 0.1}, nil
}

func TestPruneSchema(t *testing.T) {
	components := &openapi3.Components{
		Schemas: openapi3.Schemas{
			"User": {
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"object"},
				},
			},
		},
	}

	spec := &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Components: components,
		Paths: openapi3.NewPaths(),
	}

	spec.Paths.Set("/users", &openapi3.PathItem{
		Get: &openapi3.Operation{
			Summary:     "Get users",
			Description: "Get user info",
		},
	})
	spec.Paths.Set("/users/{id}", &openapi3.PathItem{
		Get: &openapi3.Operation{
			Summary:     "Get user by id",
			Description: "Get user by id",
		},
	})
	spec.Paths.Set("/login", &openapi3.PathItem{
		Post: &openapi3.Operation{
			Summary:     "Post login",
			Description: "User login",
		},
	})
	spec.Paths.Set("/posts", &openapi3.PathItem{
		Get: &openapi3.Operation{
			Summary:     "Get posts",
			Description: "Get posts",
		},
	})
	spec.Paths.Set("/empty", &openapi3.PathItem{
		Get: &openapi3.Operation{
			// Empty summary and description
		},
	})

	aiClient := &mockAIClient{}

	tests := []struct {
		name     string
		intent   string
		maxN     int
		expected []string // expected paths
	}{
		{
			name:   "Prune max 2 endpoints related to users",
			intent: "find user",
			maxN:   2,
			expected: []string{
				"/users",
				"/users/{id}",
			},
		},
		{
			name:   "Prune max 1 endpoint",
			intent: "find user",
			maxN:   1,
			expected: []string{
				"/users",
			},
		},
		{
			name:   "Bounds check: maxN higher than total endpoints",
			intent: "find user",
			maxN:   10,
			expected: []string{
				"/users",
				"/users/{id}",
				"/login",
				"/posts",
				"/empty",
			}, // Should return original paths
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pruned, err := PruneSchema(context.Background(), spec, tt.intent, tt.maxN, aiClient)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Verify Components were preserved
			if pruned.Components == nil {
				t.Fatalf("expected components to be preserved")
			}
			if pruned.Components.Schemas["User"] == nil {
				t.Errorf("expected User schema component to be preserved")
			}

			// If maxN is greater than or equal to total operations (5), we return the original spec directly.
			if tt.maxN >= 5 {
				if pruned != spec {
					t.Errorf("expected original spec to be returned for maxN >= total endpoints")
				}
				if len(pruned.Paths.Map()) != 5 {
					t.Errorf("expected 5 paths, got %d", len(pruned.Paths.Map()))
				}
				return
			}

			if len(pruned.Paths.Map()) > tt.maxN {
				t.Errorf("expected max %d endpoints, got %d", tt.maxN, len(pruned.Paths.Map()))
			}

			for _, expPath := range tt.expected {
				if pruned.Paths.Value(expPath) == nil {
					t.Errorf("expected path %s to be in pruned schema", expPath)
				}
			}
		})
	}
}
