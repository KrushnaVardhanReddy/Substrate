package schema

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/ai"
	"github.com/getkin/kin-openapi/openapi3"
)

type ScoredPath struct {
	Path      string
	Method    string
	Operation *openapi3.Operation
	Score     float64
}

// PruneSchema takes an OpenAPI spec and returns a new one containing only the top N endpoints
// that are semantically similar to the intent string.
func PruneSchema(ctx context.Context, spec *openapi3.T, intent string, maxN int, aiClient ai.AIClient) (*openapi3.T, error) {
	intentEmbedding, err := aiClient.EmbedText(ctx, intent)
	if err != nil {
		return nil, fmt.Errorf("failed to embed intent: %w", err)
	}

	var scoredPaths []ScoredPath

	for path, pathItem := range spec.Paths.Map() {
		operations := map[string]*openapi3.Operation{
			"GET":     pathItem.Get,
			"PUT":     pathItem.Put,
			"POST":    pathItem.Post,
			"DELETE":  pathItem.Delete,
			"OPTIONS": pathItem.Options,
			"HEAD":    pathItem.Head,
			"PATCH":   pathItem.Patch,
			"TRACE":   pathItem.Trace,
		}

		for method, operation := range operations {
			if operation == nil {
				continue
			}

			summary := operation.Summary
			if summary == "" {
				summary = ""
			}
			description := operation.Description
			if description == "" {
				description = ""
			}

			textToEmbed := strings.TrimSpace(fmt.Sprintf("%s %s %s", path, summary, description))
			opEmbedding, err := aiClient.EmbedText(ctx, textToEmbed)
			if err != nil {
				return nil, fmt.Errorf("failed to embed operation %s %s: %w", method, path, err)
			}

			score := ai.CosineSimilarity(intentEmbedding, opEmbedding)
			scoredPaths = append(scoredPaths, ScoredPath{
				Path:      path,
				Method:    method,
				Operation: operation,
				Score:     score,
			})
		}
	}

	sort.Slice(scoredPaths, func(i, j int) bool {
		return scoredPaths[i].Score > scoredPaths[j].Score
	})

	if maxN <= 0 {
		maxN = 5
	}

	// Bounds check: if maxN >= total paths, just return the original spec directly to avoid unnecessary work.
	// But actually, maxN is for total operations/endpoints, but the spec says "return top-N paths".
	// Let's interpret "top-N endpoints".
	if maxN >= len(scoredPaths) {
		return spec, nil
	}

	topPaths := scoredPaths[:maxN]

	// Rebuild a new openapi3.T struct with only those paths
	prunedSpec := &openapi3.T{
		OpenAPI:      spec.OpenAPI,
		Info:         spec.Info,
		Servers:      spec.Servers,
		Components:   spec.Components, // verbatim preservation of Components
		Security:     spec.Security,
		Tags:         spec.Tags,
		ExternalDocs: spec.ExternalDocs,
		Paths:        openapi3.NewPaths(),
	}

	for _, tp := range topPaths {
		pathItem := prunedSpec.Paths.Value(tp.Path)
		if pathItem == nil {
			pathItem = &openapi3.PathItem{}
			prunedSpec.Paths.Set(tp.Path, pathItem)
		}

		switch tp.Method {
		case "GET":
			pathItem.Get = tp.Operation
		case "PUT":
			pathItem.Put = tp.Operation
		case "POST":
			pathItem.Post = tp.Operation
		case "DELETE":
			pathItem.Delete = tp.Operation
		case "OPTIONS":
			pathItem.Options = tp.Operation
		case "HEAD":
			pathItem.Head = tp.Operation
		case "PATCH":
			pathItem.Patch = tp.Operation
		case "TRACE":
			pathItem.Trace = tp.Operation
		}
	}

	return prunedSpec, nil
}
