// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package ai

import (
	"context"
	"testing"

	"github.com/sashabaranov/go-openai"
)

type mockGitHubClient struct {
	prURL string
	err   error
}

func (m *mockGitHubClient) CreateDraftPR(ctx context.Context, owner, repo, branch, patch, title, body string) (string, error) {
	return m.prURL, m.err
}

type mockLLMClient struct {
	resp openai.ChatCompletionResponse
	err  error
}

func (m *mockLLMClient) CreateChatCompletion(ctx context.Context, request openai.ChatCompletionRequest) (openai.ChatCompletionResponse, error) {
	return m.resp, m.err
}

func TestGenerateAgentFixPR(t *testing.T) {
	mockLLM := &mockLLMClient{
		resp: openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Content: "```diff\n- old\n+ new\n```",
					},
				},
			},
		},
	}

	mockGithub := &mockGitHubClient{
		prURL: "https://github.com/test/repo/pull/1",
	}

	prURL, err := GenerateAgentFixPR(
		context.Background(),
		mockLLM,
		mockGithub,
		"repo",
		"owner",
		"test_tool",
		"removed parameter `database`",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if prURL != "https://github.com/test/repo/pull/1" {
		t.Fatalf("expected prURL to be https://github.com/test/repo/pull/1, got %v", prURL)
	}
}

type mockAgentRegistryClient struct {
	agents []AgentConsumer
	err    error
}

func (m *mockAgentRegistryClient) GetAgentsByTool(ctx context.Context, toolName string) ([]AgentConsumer, error) {
	return m.agents, m.err
}

func TestFixAgentsForTool(t *testing.T) {
	mockLLM := &mockLLMClient{
		resp: openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Content: "```diff\n- old\n+ new\n```",
					},
				},
			},
		},
	}

	mockGithub := &mockGitHubClient{
		prURL: "https://github.com/test/repo/pull/1",
	}

	mockRegistry := &mockAgentRegistryClient{
		agents: []AgentConsumer{
			{RepoName: "agent-repo", Owner: "agent-owner"},
		},
	}

	err := FixAgentsForTool(
		context.Background(),
		mockLLM,
		mockGithub,
		mockRegistry,
		"test_tool",
		"removed parameter `database`",
	)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
