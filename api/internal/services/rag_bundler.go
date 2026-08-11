// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

type RAGBundler struct {
	store db.Store
}

func NewRAGBundler(store db.Store) *RAGBundler {
	return &RAGBundler{store: store}
}

func (s *RAGBundler) GetContextBundle(ctx context.Context, org, repo string) (string, error) {
	targetFullName := fmt.Sprintf("%s/%s", org, repo)

	// Fetch the dependency graph for the org
	edges, err := s.store.GetDependencyGraph(ctx, org)
	if err != nil {
		// Log or wrap error as needed, but for now just return it
		return "", fmt.Errorf("failed to get dependency graph: %w", err)
	}

	// Identify direct upstreams and downstreams
	var upstreams []string // These are providers to the target consumer
	var downstreams []string // These are consumers of the target provider

	for _, edge := range edges {
		if edge.ConsumerFullName == targetFullName {
			upstreams = append(upstreams, edge.ProviderFullName)
		}
		if edge.ProviderFullName == targetFullName {
			downstreams = append(downstreams, edge.ConsumerFullName)
		}
	}

	var sb strings.Builder

	// Target Repo
	targetSchema, err := s.store.GetSchema(ctx, org, repo)
	if err != nil {
		// Even if no deps, we want the target schema.
		// If the target schema itself is missing, maybe handle it gracefully or return an error.
		return "", fmt.Errorf("failed to get target schema: %w", err)
	}

	sb.WriteString(fmt.Sprintf("# Target Service: %s\n```json\n%s\n```\n\n", repo, minifyJSON(targetSchema)))

	// Upstream Dependencies (Parents)
	for _, upFullName := range upstreams {
		parts := strings.Split(upFullName, "/")
		if len(parts) != 2 {
			continue
		}
		upOrg, upRepo := parts[0], parts[1]
		schema, err := s.store.GetSchema(ctx, upOrg, upRepo)
		if err != nil {
			continue // Skip if we can't get the schema
		}
		sb.WriteString(fmt.Sprintf("# Upstream Dependency: %s\n```json\n%s\n```\n\n", upRepo, minifyJSON(schema)))
	}

	// Downstream Consumers (Children)
	for _, downFullName := range downstreams {
		parts := strings.Split(downFullName, "/")
		if len(parts) != 2 {
			continue
		}
		downOrg, downRepo := parts[0], parts[1]
		schema, err := s.store.GetSchema(ctx, downOrg, downRepo)
		if err != nil {
			continue // Skip if we can't get the schema
		}
		sb.WriteString(fmt.Sprintf("# Downstream Consumer: %s\n```json\n%s\n```\n\n", downRepo, minifyJSON(schema)))
	}

	return strings.TrimSpace(sb.String()) + "\n", nil
}

func minifyJSON(raw string) string {
	if raw == "" {
		return "{}"
	}
	var data any
	if err := json.Unmarshal([]byte(raw), &data); err != nil {
		// If it's not valid JSON, maybe just return as is, but we should strip spaces if possible.
		// A simple string replace could work, but JSON unmarshal/marshal is safer for syntax.
		return raw
	}
	minified, err := json.Marshal(data)
	if err != nil {
		return raw
	}
	return string(minified)
}
