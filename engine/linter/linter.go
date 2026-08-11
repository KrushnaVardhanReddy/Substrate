// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package linter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/pkg/ai"
	"github.com/sashabaranov/go-openai"
)

type Issue struct {
	Rule        string `json:"rule"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

type AnalysisResponse struct {
	Score  int     `json:"score"`
	Issues []Issue `json:"issues"`
}

// FallbackMockResponse is the deterministic fallback response returned when AI features are invoked without a configured base URL.
const FallbackMockResponse = `{
  "score": 90,
  "issues": [
    {
      "rule": "FALLBACK_MODE",
      "path": "/",
      "description": "This is a deterministic fallback response because SUBSTRATE_AI_BASE_URL is unset."
    }
  ]
}`

func Analyze(ctx context.Context, schema []byte) (int, []Issue, error) {
	baseURL := viper.GetString("SUBSTRATE_AI_BASE_URL")
	apiKey := viper.GetString("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = viper.GetString("SUBSTRATE_AI_API_KEY")
	}

	if baseURL == "" && apiKey == "" {
		// Provide deterministic fallback mock response
		var resp AnalysisResponse
		if err := json.Unmarshal([]byte(FallbackMockResponse), &resp); err != nil {
			return 0, nil, err
		}
		return resp.Score, resp.Issues, nil
	}

	client, err := ai.NewAIClient()
	if err != nil {
		return 0, nil, fmt.Errorf("failed to initialize AI client: %w", err)
	}

	systemPrompt := `You are an API Design Linter. Review the provided OpenAPI schema and score it from 0 to 100 based on standard REST anti-patterns.
Specifically look for:
1. Endpoints with >10 parameters.
2. Non-descriptive field names (e.g. data1, flag).
3. Missing $ref reuse for identical structures.

Output your response as JSON in the following format, and nothing else:
{
  "score": 85,
  "issues": [
    {
      "rule": "TOO_MANY_PARAMETERS",
      "path": "/users GET",
      "description": "Endpoint has 12 parameters, which is >10."
    }
  ]
}`

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: string(schema),
			},
		},
		Stream: true,
	}

	// Wrap in a 1 minute timeout context
	ctx, cancel := context.WithTimeout(ctx, 1*time.Minute)
	defer cancel()

	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return 0, nil, fmt.Errorf("stream failed: %w", err)
	}
	defer stream.Close()

	var resultBuilder strings.Builder
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return 0, nil, fmt.Errorf("stream recv error: %w", err)
		}
		resultBuilder.WriteString(chunk)
	}

	respStr := extractJSON(resultBuilder.String())

	var resp AnalysisResponse
	if err := json.Unmarshal([]byte(respStr), &resp); err != nil {
		return 0, nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return resp.Score, resp.Issues, nil
}

func extractJSON(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}

	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}
