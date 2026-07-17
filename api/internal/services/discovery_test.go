package services

import (

	"testing"
)

func TestDiscoverImplicitDependencies(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected []string // we will just check the ProviderRepo
	}{
		{
			name: "Node.js dependencies",
			files: map[string]string{
				"package.json": `{"dependencies": {"pg": "^8.0.0", "mongoose": "^6.0.0", "stripe-node": "^1.0.0"}}`,
			},
			expected: []string{"infra:postgres", "infra:mongodb", "saas:stripe"},
		},
		{
			name: "Go dependencies",
			files: map[string]string{
				"go.mod": `
module myapp
require (
	github.com/lib/pq v1.10.9
	github.com/go-redis/redis/v8 v8.11.5
	github.com/twilio/twilio-go v1.0.0
)`,
			},
			expected: []string{"infra:postgres", "infra:redis", "saas:twilio"},
		},
		{
			name: "Docker Compose",
			files: map[string]string{
				"docker-compose.yml": `
services:
  db:
    image: bitnami/postgresql:15
  cache:
    image: redis:7
  broker:
    image: confluentinc/cp-kafka:latest
`,
			},
			expected: []string{"infra:postgres", "infra:redis", "infra:kafka"},
		},
		{
			name: "Empty files",
			files: map[string]string{
				"package.json": "",
				"go.mod":       "",
			},
			expected: nil,
		},
		{
			name: "No matches",
			files: map[string]string{
				"package.json": `{"dependencies": {"express": "^4.17.1"}}`,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := DiscoverImplicitDependencies(tt.files)

			var got []string
			for _, d := range deps {
				got = append(got, d.ProviderRepo)
			}

			if len(got) == 0 && len(tt.expected) == 0 {
				return
			}

			// We don't guarantee order from our map iteration, so we check if lengths match and all elements exist
			if len(got) != len(tt.expected) {
				t.Errorf("DiscoverImplicitDependencies() returned %d items, expected %d", len(got), len(tt.expected))
			}

			for _, exp := range tt.expected {
				found := false
				for _, g := range got {
					if g == exp {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("DiscoverImplicitDependencies() expected %s, not found in %v", exp, got)
				}
			}
		})
	}
}
