package discovery

import (
	"context"
	"sort"
)

// DependencyEdge represents a discovered connection between two repositories.
type DependencyEdge struct {
	SourceRepo string
	TargetRepo string
	TargetURL  string
	Confidence int
	Files      []string
}

// Aggregator combines and deduplicates discovery signals.
type Aggregator struct {
	edges    map[string]*DependencyEdge
	scanners []Scanner
}

// NewAggregator creates a new signal aggregator and registers the default scanners.
func NewAggregator() *Aggregator {
	a := &Aggregator{
		edges: make(map[string]*DependencyEdge),
	}
	// Register the package and openapi scanners
	a.scanners = append(a.scanners, NewPackageScanner())
	a.scanners = append(a.scanners, NewOpenAPIScanner())
	return a
}

// RegisterScanner adds a custom scanner to the aggregator (useful for testing or extensions)
func (a *Aggregator) RegisterScanner(s Scanner) {
	a.scanners = append(a.scanners, s)
}

// ScanAndAggregate runs all registered scanners and aggregates their results.
func (a *Aggregator) ScanAndAggregate(ctx context.Context, repo string, files map[string][]byte, urlResolver func(string) string) error {
	for _, scanner := range a.scanners {
		edges, err := scanner.Scan(ctx, repo, files, urlResolver)
		if err != nil {
			return err
		}
		for _, edge := range edges {
			a.AddEdge(edge)
		}
	}
	return nil
}

// AddEdge adds a new dependency edge or updates an existing one with higher confidence.
func (a *Aggregator) AddEdge(edge DependencyEdge) {
	key := edge.SourceRepo + "|" + edge.TargetRepo + "|" + edge.TargetURL
	if existing, ok := a.edges[key]; ok {
		// Combine confidence scores (up to max 100)
		existing.Confidence += edge.Confidence
		if existing.Confidence > 100 {
			existing.Confidence = 100
		}
		existing.Files = append(existing.Files, edge.Files...)
	} else {
		newEdge := edge
		a.edges[key] = &newEdge
	}
}

// GetEdges returns all aggregated edges with deterministically sorted files.
func (a *Aggregator) GetEdges() []DependencyEdge {
	var result []DependencyEdge

	// Sort edges deterministically based on map keys
	var keys []string
	for k := range a.edges {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		edge := a.edges[key]

		// Deduplicate files
		filesMap := make(map[string]bool)
		for _, file := range edge.Files {
			filesMap[file] = true
		}
		var uniqueFiles []string
		for file := range filesMap {
			uniqueFiles = append(uniqueFiles, file)
		}

		// Sort the files for deterministic output
		sort.Strings(uniqueFiles)
		edge.Files = uniqueFiles

		result = append(result, *edge)
	}
	return result
}

// Scanner defines the interface for dependency discovery scanners.
type Scanner interface {
	// Scan analyzes repository files and returns discovered dependency edges.
	Scan(ctx context.Context, repo string, files map[string][]byte, urlResolver func(string) string) ([]DependencyEdge, error)
}
