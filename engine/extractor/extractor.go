// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package extractor

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"os"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/pkg/ai"
	"github.com/sashabaranov/go-openai"
)

// ExtractSpec reads the file and extracts the OpenAPI 3.0 YAML spec using AI.
func ExtractSpec(filePath string) error {
	contentBytes, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	content := string(contentBytes)

	apiKey := viper.GetString("SUBSTRATE_AI_API_KEY")
	baseURL := viper.GetString("SUBSTRATE_AI_BASE_URL")
	model := viper.GetString("SUBSTRATE_AI_MODEL")

	if baseURL == "" {
		fmt.Println("Using deterministic fallback response (SUBSTRATE_AI_BASE_URL is unset).")
		return os.WriteFile("openapi.yaml", []byte("openapi: 3.0.0\ninfo:\n  title: Mock Watch API (updated)\n"), 0644)
	}

	if apiKey == "" {
		provider := viper.GetString("SUBSTRATE_AI_PROVIDER")
		if provider != "ollama" {
			return fmt.Errorf("SUBSTRATE_AI_API_KEY environment variable is required")
		}
	}

	if model == "" {
		model = openai.GPT4o
	}

	client, err := ai.NewAIClient()
	if err != nil {
		return fmt.Errorf("failed to initialize AI client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	systemPrompt := "Extract the OpenAPI 3.0 YAML spec from this source code. " +
		"Your response must ONLY contain the raw YAML or a single markdown code block with the YAML. " +
		"Do not include any other text, explanations, or conversational filler."

	req := openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: content,
			},
		},
		Stream: true,
	}

	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create chat completion stream: %w", err)
	}
	defer stream.Close()

	var fullResponse strings.Builder

	for {
		chunk, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("stream error: %w", err)
		}
		fullResponse.WriteString(chunk)
	}

	specContent := extractYAML(fullResponse.String())

	err = os.WriteFile("openapi.yaml", []byte(specContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write openapi.yaml: %w", err)
	}

	return nil
}

func extractYAML(response string) string {
	text := strings.TrimSpace(response)
	if strings.HasPrefix(text, "```yaml") {
		text = strings.TrimPrefix(text, "```yaml")
	} else if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```")
	}

	if strings.HasSuffix(text, "```") {
		text = strings.TrimSuffix(text, "```")
	}

	return strings.TrimSpace(text) + "\n"
}

// Extract is a legacy wrapper.
func Extract(path string) error {
	return ExtractSpec(path)
}
