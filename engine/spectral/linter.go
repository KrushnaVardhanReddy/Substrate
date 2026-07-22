package spectral

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/pkg/ai"
	"github.com/sashabaranov/go-openai"
)

type Linter struct {
	client ai.AIClient
}

func NewLinter(client ai.AIClient) *Linter {
	return &Linter{client: client}
}

type Violation struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

func (l *Linter) LintDiff(diff string, rules []string) ([]Violation, error) {
	if len(rules) == 0 {
		return nil, nil
	}

	rulesText := "- " + strings.Join(rules, "\n- ")
	prompt := fmt.Sprintf(`You are an API governance linter. Review the following OpenAPI diff against these plain-English rules:

Rules:
%s

Diff:
%s

Analyze the diff and identify any violations of the rules. For each violation, explain clearly what the violation is and which rule was broken.
Return a list of violations. If there are none, simply output "NO_VIOLATIONS".`, rulesText, diff)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	stream, err := l.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are an API governance linter. You only evaluate the provided diff against the provided rules.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion stream: %w", err)
	}
	defer stream.Close()

	var result strings.Builder
	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" || strings.Contains(err.Error(), "EOF") {
				break
			}
			return nil, fmt.Errorf("error reading stream: %w", err)
		}
		result.WriteString(chunk)
	}

	output := strings.TrimSpace(result.String())
	if output == "NO_VIOLATIONS" {
		return nil, nil
	}

	return []Violation{
		{
			Rule:    "AI Linter Findings",
			Message: output,
		},
	}, nil
}
