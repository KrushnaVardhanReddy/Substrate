package ai

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"os"
	"strings"

	"github.com/sashabaranov/go-openai"
)

var (
	ErrMissingAPIKey = errors.New("SUBSTRATE_AI_API_KEY environment variable is required")
)

// FallbackOpenAPI is the deterministic mock response returned if SUBSTRATE_AI_BASE_URL is unset
// and SUBSTRATE_AI_API_KEY is unset. Wait, memory says "must implement deterministic fallback mock responses when SUBSTRATE_AI_BASE_URL is unset"
const FallbackOpenAPI = `openapi: 3.0.0
info:
  title: Mock API
  version: 1.0.0
paths:
  /hello:
    get:
      summary: Say Hello
      responses:
        '200':
          description: OK
`

func RunArchitect(in io.Reader, out io.Writer) error {
	fmt.Fprint(out, "Describe the API you want to build (e.g., 'I need a blog API with posts and comments'): ")

	reader := bufio.NewReader(in)
	userInput, err := reader.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return errors.New("input closed unexpectedly")
		}
		return fmt.Errorf("failed to read input: %w", err)
	}

	userInput = strings.TrimSpace(userInput)
	if userInput == "" {
		return errors.New("input cannot be empty")
	}

	apiKey := viper.GetString("SUBSTRATE_AI_API_KEY")
	baseURL := viper.GetString("SUBSTRATE_AI_BASE_URL")
	model := viper.GetString("SUBSTRATE_AI_MODEL")

	// Apply Fallback behavior: "must implement deterministic fallback mock responses when SUBSTRATE_AI_BASE_URL is unset"
	// However, if the API key is set, maybe it's just meant for local OpenAI? The memory specifically says "when SUBSTRATE_AI_BASE_URL is unset" for AI endpoints...
	// Let's check exactly: "AI endpoints in the Substrate API expect configuration via SUBSTRATE_AI_BASE_URL, SUBSTRATE_AI_API_KEY, and SUBSTRATE_AI_MODEL environment variables, and must implement deterministic fallback mock responses when SUBSTRATE_AI_BASE_URL is unset."
	// Note: It's CLI, not API. But the prompt says "AI endpoints in the Substrate API...". Still, for the CLI we also need it for tests.
	if baseURL == "" {
		fmt.Fprintln(out, "Using deterministic fallback response (SUBSTRATE_AI_BASE_URL is unset).")
		err = os.WriteFile("openapi.yaml", []byte(FallbackOpenAPI), 0644)
		if err != nil {
			return fmt.Errorf("failed to write openapi.yaml: %w", err)
		}
		return nil
	}

	if apiKey == "" {
		provider := viper.GetString("SUBSTRATE_AI_PROVIDER")
		if provider != "ollama" {
			return ErrMissingAPIKey
		}
	}

	if model == "" {
		model = openai.GPT4o
	}

	client, err := NewAIClient()
	if err != nil {
		return fmt.Errorf("failed to initialize AI client: %w", err)
	}
	ctx := context.Background()

	systemPrompt := "You are Substrate AI Architect. The user will describe an API they want. " +
		"You must generate a valid OpenAPI 3.0 YAML specification for it. " +
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
				Content: userInput,
			},
		},
		Stream: true,
	}

	stream, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to create chat completion stream: %w", err)
	}
	defer stream.Close()

	fmt.Fprintln(out, "\nGenerating your API spec...")

	var fullResponse strings.Builder

	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("stream error: %w", err)
		}

		fullResponse.WriteString(chunk)
		fmt.Fprint(out, chunk)
	}

	fmt.Fprintln(out) // Newline after stream finishes

	specContent := extractYAML(fullResponse.String())

	err = os.WriteFile("openapi.yaml", []byte(specContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write openapi.yaml: %w", err)
	}

	return nil
}

func extractYAML(response string) string {
	// Simple extractor: if it starts with ```yaml and ends with ```, strip them.
	// Otherwise return the string as is.
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
