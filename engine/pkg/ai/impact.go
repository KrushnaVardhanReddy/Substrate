package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sashabaranov/go-openai"
)

type Change struct {
	RuleID      string `json:"rule_id"`
	Path        string `json:"path"`
	Description string `json:"description"`
}

type Repo struct {
	Name string `json:"name"`
}

func GenerateImpactSummary(ctx context.Context, llmClient AIClient, changes []Change, consumers []Repo) (string, error) {
	if len(changes) == 0 {
		return "No breaking changes detected.", nil
	}

	changesJSON, _ := json.Marshal(changes)
	consumersJSON, _ := json.Marshal(consumers)

	prompt := fmt.Sprintf(`You are an AI assistant analyzing API breaking changes.
Given the following breaking changes:
%s
And the following affected downstream consumers:
%s
Provide a concise risk narrative summarizing the blast radius impact in plain English.
Limit your response to 200 tokens.`, changesJSON, consumersJSON)

	req := openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a concise API impact analyzer.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens: 200,
		Stream:    true,
	}

	stream, err := llmClient.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to call LLM: %w", err)
	}
	defer stream.Close()

	var summary strings.Builder
	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return "", fmt.Errorf("stream error: %w", err)
		}
		summary.WriteString(chunk)
	}

	return strings.TrimSpace(summary.String()), nil
}
