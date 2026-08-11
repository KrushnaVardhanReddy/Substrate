// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package discovery

import (
	"context"
	"reflect"
	"testing"
)

func TestAggregator(t *testing.T) {
	tests := []struct {
		name     string
		edges    []DependencyEdge
		expected []DependencyEdge
	}{
		{
			name: "add and combine edges",
			edges: []DependencyEdge{
				{
					SourceRepo: "myorg/frontend",
					TargetRepo: "myorg/backend",
					Confidence: 40,
					Files:      []string{"package.json"},
				},
				{
					SourceRepo: "myorg/frontend",
					TargetRepo: "myorg/backend",
					Confidence: 30,
					Files:      []string{".env.example"},
				},
			},
			expected: []DependencyEdge{
				{
					SourceRepo: "myorg/frontend",
					TargetRepo: "myorg/backend",
					Confidence: 70,
					Files:      []string{".env.example", "package.json"},
				},
			},
		},
		{
			name: "cap confidence at 100",
			edges: []DependencyEdge{
				{
					SourceRepo: "myorg/frontend",
					TargetRepo: "myorg/backend",
					Confidence: 60,
					Files:      []string{"file1"},
				},
				{
					SourceRepo: "myorg/frontend",
					TargetRepo: "myorg/backend",
					Confidence: 50,
					Files:      []string{"file2"},
				},
			},
			expected: []DependencyEdge{
				{
					SourceRepo: "myorg/frontend",
					TargetRepo: "myorg/backend",
					Confidence: 100,
					Files:      []string{"file1", "file2"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAggregator()
			for _, edge := range tt.edges {
				a.AddEdge(edge)
			}

			edges := a.GetEdges()
			if len(edges) != len(tt.expected) {
				t.Fatalf("expected %d edges, got %d", len(tt.expected), len(edges))
			}

			for i, expectedEdge := range tt.expected {
				if edges[i].Confidence != expectedEdge.Confidence {
					t.Errorf("expected confidence %d, got %d", expectedEdge.Confidence, edges[i].Confidence)
				}
				if len(edges[i].Files) != len(expectedEdge.Files) {
					t.Fatalf("expected %d files, got %d", len(expectedEdge.Files), len(edges[i].Files))
				}
				for j, file := range expectedEdge.Files {
					if edges[i].Files[j] != file {
						t.Errorf("expected file %s, got %s", file, edges[i].Files[j])
					}
				}
			}
		})
	}
}

// mockScanner for testing ScanAndAggregate
type mockScanner struct {
	edges []DependencyEdge
	err   error
}

func (m *mockScanner) Scan(ctx context.Context, repo string, files map[string][]byte, urlResolver func(string) string) ([]DependencyEdge, error) {
	return m.edges, m.err
}

func TestAggregatorScanAndAggregate(t *testing.T) {
	a := NewAggregator()
	// Clear default scanners for deterministic testing
	a.scanners = []Scanner{}

	a.RegisterScanner(&mockScanner{
		edges: []DependencyEdge{
			{
				SourceRepo: "repoA",
				TargetRepo: "repoB",
				Confidence: 50,
				Files:      []string{"mockFile"},
			},
		},
	})

	err := a.ScanAndAggregate(context.Background(), "repoA", map[string][]byte{}, func(string) string { return "" })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	edges := a.GetEdges()
	if len(edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(edges))
	}
	if edges[0].SourceRepo != "repoA" || edges[0].TargetRepo != "repoB" || edges[0].Confidence != 50 {
		t.Errorf("unexpected edge contents: %+v", edges[0])
	}
}

func TestAggregateSignals(t *testing.T) {
	tests := []struct {
		name  string
		edges []Edge
		want  []Edge
	}{
		{
			name: "map producers and consumers via shared Topic names",
			edges: []Edge{
				{
					SourceRepo: "myorg/orders-service",
					TargetRepo: "myorg/orders-service",
					Confidence: 20,
					Signal:     "kafka_topic:orders.v2.order_created",
				},
				{
					SourceRepo: "myorg/payments-service",
					TargetRepo: "myorg/payments-service",
					Confidence: 20,
					Signal:     "kafka_topic:orders.v2.order_created",
				},
			},
			want: []Edge{
				{
					SourceRepo: "myorg/orders-service",
					TargetRepo: "myorg/payments-service",
					Confidence: 20,
					Signal:     "kafka_topic_shared",
				},
				{
					SourceRepo: "myorg/payments-service",
					TargetRepo: "myorg/orders-service",
					Confidence: 20,
					Signal:     "kafka_topic_shared",
				},
			},
		},
		{
			name: "passthrough other edges",
			edges: []Edge{
				{
					SourceRepo: "myorg/a",
					TargetRepo: "myorg/b",
					Confidence: 35,
					Signal:     "asyncapi_$ref",
				},
			},
			want: []Edge{
				{
					SourceRepo: "myorg/a",
					TargetRepo: "myorg/b",
					Confidence: 35,
					Signal:     "asyncapi_$ref",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AggregateSignals(tt.edges)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AggregateSignals() = %v, want %v", got, tt.want)
			}
		})
	}
}
