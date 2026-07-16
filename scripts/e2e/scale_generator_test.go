//go:build ignore

package main

import (
	"testing"
)

func TestScaleGenerator_1000Nodes(t *testing.T) {
	edges := GenerateScaleGraph(1000, 3000)

	// Assert: at least 2999 edges
	if len(edges) < 2999 {
		t.Fatalf("expected >= 2999 edges, got %d", len(edges))
	}

	// Assert: exactly 1000 unique nodes
	nodes := map[string]bool{}
	for _, e := range edges {
		nodes[e.Provider] = true
		nodes[e.Consumer] = true
	}
	if len(nodes) != 1000 {
		t.Errorf("expected 1000 unique nodes, got %d", len(nodes))
	}

	// Assert: no duplicate edges
	seen := map[string]bool{}
	for _, e := range edges {
		key := e.Provider + ">" + e.Consumer
		if seen[key] {
			t.Errorf("duplicate edge: %s", key)
		}
		seen[key] = true
	}

	// Assert: no cycles via DFS from node-0 (forward-only check)
	adj := map[string][]string{}
	for _, e := range edges {
		adj[e.Provider] = append(adj[e.Provider], e.Consumer)
	}

	visited := map[string]bool{}
	inPath := map[string]bool{}
	var dfs func(node string) bool
	dfs = func(node string) bool {
		if inPath[node] {
			return true
		}
		if visited[node] {
			return false
		}
		visited[node] = true
		inPath[node] = true
		for _, next := range adj[node] {
			if dfs(next) {
				return true
			}
		}
		inPath[node] = false
		return false
	}
	if dfs("node-0") {
		t.Error("cycle detected in generated graph")
	}
}
