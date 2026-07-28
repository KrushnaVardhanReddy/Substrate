package ai

import (
	"context"
	"fmt"
	"math"

	"github.com/sashabaranov/go-openai"
	"github.com/spf13/viper"
)

type AIClient interface {
	EmbedText(ctx context.Context, text string) ([]float64, error)
}

type DefaultAIClient struct{}

func NewClient() AIClient {
	return &DefaultAIClient{}
}

func (c *DefaultAIClient) EmbedText(ctx context.Context, text string) ([]float64, error) {
	return EmbedText(ctx, text)
}

// EmbedText gets an embedding for the given text using the go-openai client.
func EmbedText(ctx context.Context, text string) ([]float64, error) {
	apiKey := viper.GetString("SUBSTRATE_AI_API_KEY")
	if apiKey == "" {
		apiKey = viper.GetString("OPENAI_API_KEY")
	}

	if apiKey == "" {
		return nil, fmt.Errorf("openai API key is required")
	}

	config := openai.DefaultConfig(apiKey)

	baseURL := viper.GetString("SUBSTRATE_AI_BASE_URL")
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(config)

	req := openai.EmbeddingRequest{
		Input: []string{text},
		Model: openai.SmallEmbedding3,
	}

	resp, err := client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create embeddings: %w", err)
	}

	if len(resp.Data) > 0 {
		// Convert []float32 to []float64
		embedding32 := resp.Data[0].Embedding
		embedding64 := make([]float64, len(embedding32))
		for i, v := range embedding32 {
			embedding64[i] = float64(v)
		}
		return embedding64, nil
	}

	return nil, fmt.Errorf("no embedding returned")
}

func CosineSimilarity(a, b []float64) float64 {
	var dotProduct float64
	var normA float64
	var normB float64

	for i := 0; i < len(a) && i < len(b); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA == 0 || normB == 0 {
		return 0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
