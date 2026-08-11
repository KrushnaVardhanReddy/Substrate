// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package services

import (
	"context"
	"strings"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

// mockStore is a simplified mock for db.Store
type mockStore struct {
	db.Store // embed to satisfy interface, panic on unimplemented methods
	edges    []db.DependencyEdge
	schemas  map[string]string
}

func (m *mockStore) GetDependencyGraph(ctx context.Context, orgName string) ([]db.DependencyEdge, error) {
	return m.edges, nil
}

func (m *mockStore) GetSchema(ctx context.Context, org, repo string) (string, error) {
	schema, ok := m.schemas[org+"/"+repo]
	if !ok {
		return "", db.ErrNotFound
	}
	return schema, nil
}

func TestRAGBundler_GetContextBundle(t *testing.T) {
	tests := []struct {
		name        string
		org         string
		repo        string
		edges       []db.DependencyEdge
		schemas     map[string]string
		wantContain []string
		wantErr     bool
	}{
		{
			name: "Target repo with upstreams and downstreams",
			org:  "testorg",
			repo: "target-api",
			edges: []db.DependencyEdge{
				{ConsumerFullName: "testorg/target-api", ProviderFullName: "testorg/upstream-api"},
				{ConsumerFullName: "testorg/downstream-api", ProviderFullName: "testorg/target-api"},
			},
			schemas: map[string]string{
				"testorg/target-api":     `{ "openapi": "3.0.0", "info": { "title": "Target" } }`,
				"testorg/upstream-api":   `{ "openapi": "3.0.0", "info": { "title": "Upstream" } }`,
				"testorg/downstream-api": `{ "openapi": "3.0.0", "info": { "title": "Downstream" } }`,
			},
			wantContain: []string{
				"# Target Service: target-api",
				`{"info":{"title":"Target"},"openapi":"3.0.0"}`,
				"# Upstream Dependency: upstream-api",
				`{"info":{"title":"Upstream"},"openapi":"3.0.0"}`,
				"# Downstream Consumer: downstream-api",
				`{"info":{"title":"Downstream"},"openapi":"3.0.0"}`,
			},
		},
		{
			name: "Target repo with no dependencies",
			org:  "testorg",
			repo: "lonely-api",
			edges: []db.DependencyEdge{},
			schemas: map[string]string{
				"testorg/lonely-api": `{ "openapi": "3.0.0" }`,
			},
			wantContain: []string{
				"# Target Service: lonely-api",
				`{"openapi":"3.0.0"}`,
			},
		},
		{
			name: "Missing target schema returns error",
			org:  "testorg",
			repo: "missing-api",
			edges: []db.DependencyEdge{},
			schemas: map[string]string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStore{
				edges:   tt.edges,
				schemas: tt.schemas,
			}
			bundler := NewRAGBundler(store)
			got, err := bundler.GetContextBundle(context.Background(), tt.org, tt.repo)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetContextBundle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			for _, want := range tt.wantContain {
				if !strings.Contains(got, want) {
					t.Errorf("GetContextBundle() output missing expected string: %s\nGot:\n%s", want, got)
				}
			}

			// Verify minification - no extra spaces or newlines in the JSON part
			// By checking that `{ ` doesn't exist (unless it's part of the raw string and unmarshal failed)
			// A quick check to see if the JSON is compact.
			if strings.Contains(got, "{ \"openapi\"") {
				t.Errorf("GetContextBundle() output doesn't seem minified: %s", got)
			}
		})
	}
}
