package ai

import (
	"context"
	"errors"
	"fmt"
	"github.com/spf13/viper"

	"github.com/sashabaranov/go-openai"
)

var (
	ErrUnsupportedProvider = errors.New("unsupported AI provider")
)

type AIChatCompletionStream interface {
	Recv() (string, error)
	Close() error
}

type AIClient interface {
	CreateChatCompletionStream(ctx context.Context, request openai.ChatCompletionRequest) (AIChatCompletionStream, error)
}

// openaiStreamWrapper wraps the go-openai stream to match our interface
type openaiStreamWrapper struct {
	stream *openai.ChatCompletionStream
}

func (w *openaiStreamWrapper) Recv() (string, error) {
	resp, err := w.stream.Recv()
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Delta.Content, nil
}

func (w *openaiStreamWrapper) Close() error {
	w.stream.Close()
	return nil
}

type openaiClientWrapper struct {
	client *openai.Client
}

func (c *openaiClientWrapper) CreateChatCompletionStream(ctx context.Context, request openai.ChatCompletionRequest) (AIChatCompletionStream, error) {
	stream, err := c.client.CreateChatCompletionStream(ctx, request)
	if err != nil {
		return nil, err
	}
	return &openaiStreamWrapper{stream: stream}, nil
}

func NewAIClient() (AIClient, error) {
	provider := viper.GetString("SUBSTRATE_AI_PROVIDER")
	apiKey := viper.GetString("SUBSTRATE_AI_API_KEY")
	baseURL := viper.GetString("SUBSTRATE_AI_BASE_URL")

	if provider == "" || provider == "openai" {
		config := openai.DefaultConfig(apiKey)
		if baseURL != "" {
			config.BaseURL = baseURL
		}
		return &openaiClientWrapper{client: openai.NewClientWithConfig(config)}, nil
	}

	if provider == "azure" {
		config := openai.DefaultAzureConfig(apiKey, baseURL)
		return &openaiClientWrapper{client: openai.NewClientWithConfig(config)}, nil
	}

	if provider == "ollama" {
		config := openai.DefaultConfig("ollama")
		if baseURL != "" {
			config.BaseURL = baseURL
		} else {
			config.BaseURL = "http://localhost:11434/v1"
		}
		return &openaiClientWrapper{client: openai.NewClientWithConfig(config)}, nil
	}

	if provider == "bedrock" {
		return NewBedrockAIClient(baseURL, apiKey), nil
	}

	return nil, fmt.Errorf("%w: %s", ErrUnsupportedProvider, provider)
}
