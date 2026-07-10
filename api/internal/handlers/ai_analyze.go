package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

type AIAnalyzeRequest struct {
	Org            string `json:"org"`
	CurrentSchema  string `json:"current_schema"`
	ProposedSchema string `json:"proposed_schema"`
	SchemaType     string `json:"schema_type"` // "openapi" | "graphql" | "sql"
}

type SSEEvent struct {
	Type     string `json:"type"`               // "thinking" | "finding" | "fix" | "done" | "error"
	Content  string `json:"content,omitempty"`
	Severity string `json:"severity,omitempty"` // "BREAKING" | "WARNING" | "SAFE"
	Language string `json:"language,omitempty"` // "yaml" | "sql" | "graphql"
	Code     string `json:"code,omitempty"`
}

type AIConfig struct {
	BaseURL string // SUBSTRATE_AI_BASE_URL env var
	APIKey  string // SUBSTRATE_AI_API_KEY env var
	Model   string // SUBSTRATE_AI_MODEL env var
}

func loadAIConfig() AIConfig {
	return AIConfig{
		BaseURL: os.Getenv("SUBSTRATE_AI_BASE_URL"),
		APIKey:  os.Getenv("SUBSTRATE_AI_API_KEY"),
		Model:   os.Getenv("SUBSTRATE_AI_MODEL"),
	}
}

func writeSSE(w http.ResponseWriter, f http.Flusher, event SSEEvent) error {
	data, _ := json.Marshal(event)
	_, err := fmt.Fprintf(w, "data: %s\n\n", data)
	f.Flush()
	return err
}

func mockSSEResponse(w http.ResponseWriter, f http.Flusher) {
	writeSSE(w, f, SSEEvent{Type: "thinking", Content: "Substrate AI (mock mode) — set SUBSTRATE_AI_BASE_URL to enable real AI."})
	writeSSE(w, f, SSEEvent{Type: "finding", Severity: "BREAKING", Content: "Removing a field from a response schema will break consumers that depend on it."})
	writeSSE(w, f, SSEEvent{Type: "fix", Language: "yaml", Code: "# Mark the field as deprecated instead of removing it:\nuser_id:\n  type: string\n  deprecated: true"})
	writeSSE(w, f, SSEEvent{Type: "done"})
}

const systemPrompt = `You are Substrate AI, an expert in API schema governance and contract safety.
You will be given a diff between two schema versions. Your job is to:
1. Explain in 2-3 plain-English sentences what changed and why it matters.
2. Identify the severity: BREAKING, WARNING, or SAFE.
3. Suggest the minimal safe remediation. For breaking changes, prefer deprecation over deletion.
4. Provide the exact corrected schema snippet as a fenced code block.

Be concise. Ground all claims in the diff provided. Do not hallucinate fields not present in the input.`

func AIAnalyzeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Step 1: Set SSE headers BEFORE writing anything
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)

		// Step 2: Assert http.Flusher
		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		// Step 3: Parse the JSON request body
		var req AIAnalyzeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeSSE(w, flusher, SSEEvent{Type: "error", Content: "invalid request body"})
			writeSSE(w, flusher, SSEEvent{Type: "done"})
			return
		}

		// Step 4: Read AIConfig from env
		cfg := loadAIConfig()
		if cfg.BaseURL == "" {
			mockSSEResponse(w, flusher)
			return
		}

		// Step 6: Build the user message
		userMessage := fmt.Sprintf("Schema type: %s\n\nCurrent schema:\n%s\n\nProposed schema:\n%s", req.SchemaType, req.CurrentSchema, req.ProposedSchema)

		// Step 7: Call the OpenAI-compatible chat completions endpoint
		type chatMessage struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}
		type chatRequest struct {
			Model    string        `json:"model"`
			Stream   bool          `json:"stream"`
			Messages []chatMessage `json:"messages"`
		}

		body := chatRequest{
			Model:  cfg.Model,
			Stream: true,
			Messages: []chatMessage{
				{Role: "system", Content: systemPrompt},
				{Role: "user", Content: userMessage},
			},
		}

		bodyBytes, err := json.Marshal(body)
		if err != nil {
			writeSSE(w, flusher, SSEEvent{Type: "error", Content: "failed to marshal request body"})
			writeSSE(w, flusher, SSEEvent{Type: "done"})
			return
		}

		httpReq, err := http.NewRequest("POST", cfg.BaseURL+"/chat/completions", bytes.NewBuffer(bodyBytes))
		if err != nil {
			writeSSE(w, flusher, SSEEvent{Type: "error", Content: "failed to create request"})
			writeSSE(w, flusher, SSEEvent{Type: "done"})
			return
		}

		httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
		httpReq.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(httpReq)
		if err != nil {
			writeSSE(w, flusher, SSEEvent{Type: "error", Content: "failed to call AI provider"})
			writeSSE(w, flusher, SSEEvent{Type: "done"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			writeSSE(w, flusher, SSEEvent{Type: "error", Content: fmt.Sprintf("AI provider returned status: %d", resp.StatusCode)})
			writeSSE(w, flusher, SSEEvent{Type: "done"})
			return
		}

		// Step 8: Stream the response
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" || !strings.HasPrefix(line, "data: ") {
				continue
			}

			dataStr := strings.TrimPrefix(line, "data: ")
			if dataStr == "[DONE]" {
				continue
			}

			var chunk struct {
				Choices []struct {
					Delta struct {
						Content string `json:"content"`
					} `json:"delta"`
				} `json:"choices"`
			}

			if err := json.Unmarshal([]byte(dataStr), &chunk); err != nil {
				continue // skip invalid chunks
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				writeSSE(w, flusher, SSEEvent{Type: "thinking", Content: chunk.Choices[0].Delta.Content})
			}
		}

		// Step 9: After the LLM stream ends, stream done
		writeSSE(w, flusher, SSEEvent{Type: "done"})
	}
}
