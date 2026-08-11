// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package discovery

import (
	"context"
	"encoding/json"
	"strings"
)

// PackageScanner implements Scanner for package.json and go.mod files
type PackageScanner struct{}

func NewPackageScanner() *PackageScanner {
	return &PackageScanner{}
}

func (s *PackageScanner) Scan(ctx context.Context, repo string, files map[string][]byte, urlResolver func(string) string) ([]DependencyEdge, error) {
	var edges []DependencyEdge

	// Scan package.json
	if content, ok := files["package.json"]; ok {
		edges = append(edges, s.scanPackageJSON(repo, content, urlResolver)...)
	}

	// Scan go.mod
	if content, ok := files["go.mod"]; ok {
		edges = append(edges, s.scanGoMod(repo, content, urlResolver)...)
	}

	return edges, nil
}

func (s *PackageScanner) scanPackageJSON(repo string, content []byte, urlResolver func(string) string) []DependencyEdge {
	var edges []DependencyEdge

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}

	if err := json.Unmarshal(content, &pkg); err != nil {
		return edges // Skip on error
	}

	checkDeps := func(deps map[string]string) {
		for name := range deps {
			if strings.HasPrefix(name, "@") {
				// E.g. @myorg/backend-sdk
				if targetRepo := urlResolver(name); targetRepo != "" {
					edges = append(edges, DependencyEdge{
						SourceRepo: repo,
						TargetRepo: targetRepo,
						Confidence: 40,
						Files:      []string{"package.json"},
					})
				}
			}
		}
	}

	checkDeps(pkg.Dependencies)
	checkDeps(pkg.DevDependencies)

	return edges
}

func (s *PackageScanner) scanGoMod(repo string, content []byte, urlResolver func(string) string) []DependencyEdge {
	var edges []DependencyEdge

	// A simple scan for require github.com/org/repo
	lines := strings.Split(string(content), "\n")
	inRequireBlock := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if strings.HasPrefix(line, "require (") {
			inRequireBlock = true
			continue
		}

		if inRequireBlock && line == ")" {
			inRequireBlock = false
			continue
		}

		var modulePath string
		if inRequireBlock {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				modulePath = parts[0]
			}
		} else if strings.HasPrefix(line, "require ") {
			parts := strings.Fields(line)
			if len(parts) > 1 {
				modulePath = parts[1]
			}
		}

		if modulePath != "" {
			if targetRepo := urlResolver(modulePath); targetRepo != "" {
				edges = append(edges, DependencyEdge{
					SourceRepo: repo,
					TargetRepo: targetRepo,
					Confidence: 40,
					Files:      []string{"go.mod"},
				})
			}
		}
	}

	return edges
}
