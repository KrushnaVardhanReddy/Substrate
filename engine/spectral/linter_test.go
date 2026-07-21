package spectral

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/engine/pkg/ai"
	"github.com/sashabaranov/go-openai"
)

type mockStream struct {
	chunks []string
	idx    int
}

func (m *mockStream) Recv() (string, error) {
	if m.idx >= len(m.chunks) {
		return "", errors.New("EOF")
	}
	res := m.chunks[m.idx]
	m.idx++
	return res, nil
}

func (m *mockStream) Close() error {
	return nil
}

type mockAIClient struct {
	response string
	err      error
}

func (m *mockAIClient) CreateChatCompletionStream(ctx context.Context, request openai.ChatCompletionRequest) (ai.AIChatCompletionStream, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &mockStream{chunks: []string{m.response}}, nil
}

func TestLinter_LintDiff(t *testing.T) {
	tests := []struct {
		name       string
		rules      []string
		diff       string
		aiResponse string
		aiError    error
		wantCount  int
		wantErr    bool
	}{
		{
			name:       "No rules",
			rules:      []string{},
			diff:       "+ /test",
			aiResponse: "",
			wantCount:  0,
			wantErr:    false,
		},
		{
			name:       "No violations",
			rules:      []string{"All endpoints must be camelCase"},
			diff:       "+ /camelCase",
			aiResponse: "NO_VIOLATIONS",
			wantCount:  0,
			wantErr:    false,
		},
		{
			name:       "Has violations",
			rules:      []string{"All endpoints must be camelCase"},
			diff:       "+ /snake_case",
			aiResponse: "Violation: /snake_case breaks the camelCase rule.",
			wantCount:  1,
			wantErr:    false,
		},
		{
			name:       "AI Client Error",
			rules:      []string{"rule"},
			diff:       "diff",
			aiResponse: "",
			aiError:    errors.New("ai failed"),
			wantCount:  0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockAIClient{
				response: tt.aiResponse,
				err:      tt.aiError,
			}
			linter := NewLinter(client)
			violations, err := linter.LintDiff(tt.diff, tt.rules)

			if (err != nil) != tt.wantErr {
				t.Errorf("LintDiff() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(violations) != tt.wantCount {
				t.Errorf("LintDiff() got %v violations, want %v", len(violations), tt.wantCount)
			}

			if tt.wantCount > 0 {
				if !strings.Contains(violations[0].Message, tt.aiResponse) {
					t.Errorf("LintDiff() got violation message = %v, want to contain %v", violations[0].Message, tt.aiResponse)
				}
			}
		})
	}
}
