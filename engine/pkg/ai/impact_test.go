package ai

import (
	"context"
	"io"
	"testing"

	"github.com/sashabaranov/go-openai"
)

type mockStream struct {
	chunks []string
	idx    int
}

func (m *mockStream) Recv() (string, error) {
	if m.idx >= len(m.chunks) {
		return "", io.EOF
	}
	res := m.chunks[m.idx]
	m.idx++
	return res, nil
}

func (m *mockStream) Close() error {
	return nil
}

type mockAIClient struct {
	stream AIChatCompletionStream
	err    error
}

func (m *mockAIClient) CreateChatCompletionStream(ctx context.Context, request openai.ChatCompletionRequest) (AIChatCompletionStream, error) {
	return m.stream, m.err
}

func TestGenerateImpactSummary(t *testing.T) {
	tests := []struct {
		name      string
		changes   []Change
		consumers []Repo
		mockResp  []string
		expected  string
	}{
		{
			name: "no changes",
			changes: []Change{},
			consumers: []Repo{{Name: "repo1"}},
			mockResp: []string{},
			expected: "No breaking changes detected.",
		},
		{
			name: "with changes",
			changes: []Change{
				{RuleID: "TEST", Path: "/test", Description: "removed"},
			},
			consumers: []Repo{{Name: "repo1"}},
			mockResp: []string{"Impact ", "is ", "bad."},
			expected: "Impact is bad.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &mockAIClient{
				stream: &mockStream{chunks: tt.mockResp},
			}
			got, err := GenerateImpactSummary(context.Background(), client, tt.changes, tt.consumers)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
