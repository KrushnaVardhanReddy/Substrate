package services

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
	Type     string `json:"type"` // "thinking" | "finding" | "fix" | "done" | "error"
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
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
			return
		}

		var req AIAnalyzeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeSSE(w, flusher, SSEEvent{Type: "error", Content: "invalid request body"})
			writeSSE(w, flusher, SSEEvent{Type: "done"})
			return
		}

		cfg := loadAIConfig()
		if cfg.BaseURL == "" {
			mockSSEResponse(w, flusher)
			return
		}

		userMessage := fmt.Sprintf("Schema type: %s\n\nCurrent schema:\n%s\n\nProposed schema:\n%s", req.SchemaType, req.CurrentSchema, req.ProposedSchema)

		messages := []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		}

		client := &http.Client{}

		for {
			body := chatRequest{
				Model:    cfg.Model,
				Stream:   true,
				Messages: messages,
				Tools:    mcpTools,
			}

			bodyBytes, err := json.Marshal(body)
			if err != nil {
				writeSSE(w, flusher, SSEEvent{Type: "error", Content: "failed to marshal request body"})
				break
			}

			httpReq, err := http.NewRequest("POST", cfg.BaseURL+"/chat/completions", bytes.NewBuffer(bodyBytes))
			if err != nil {
				writeSSE(w, flusher, SSEEvent{Type: "error", Content: "failed to create request"})
				break
			}

			httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
			httpReq.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(httpReq)
			if err != nil {
				writeSSE(w, flusher, SSEEvent{Type: "error", Content: "failed to call AI provider"})
				break
			}

			if resp.StatusCode != http.StatusOK {
				resp.Body.Close()
				writeSSE(w, flusher, SSEEvent{Type: "error", Content: fmt.Sprintf("AI provider returned status: %d", resp.StatusCode)})
				break
			}

			scanner := bufio.NewScanner(resp.Body)

			var pendingToolCalls []toolCall
			var assistantMessageContent string
			toolCallsMap := make(map[int]*toolCall)

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
						FinishReason *string `json:"finish_reason"`
						Delta        struct {
							Content   string `json:"content"`
							ToolCalls []struct {
								Index    int    `json:"index"`
								ID       string `json:"id"`
								Type     string `json:"type"`
								Function struct {
									Name      string `json:"name"`
									Arguments string `json:"arguments"`
								} `json:"function"`
							} `json:"tool_calls"`
						} `json:"delta"`
					} `json:"choices"`
				}

				if err := json.Unmarshal([]byte(dataStr), &chunk); err != nil {
					continue
				}

				if len(chunk.Choices) > 0 {
					choice := chunk.Choices[0]

					// Stream text to frontend as thinking
					if choice.Delta.Content != "" {
						assistantMessageContent += choice.Delta.Content
						writeSSE(w, flusher, SSEEvent{Type: "thinking", Content: choice.Delta.Content})
					}

					// Accumulate tool calls
					for _, tcDelta := range choice.Delta.ToolCalls {
						idx := tcDelta.Index
						if toolCallsMap[idx] == nil {
							toolCallsMap[idx] = &toolCall{
								ID:   tcDelta.ID,
								Type: tcDelta.Type,
							}
							toolCallsMap[idx].Function.Name = tcDelta.Function.Name
						}
						toolCallsMap[idx].Function.Arguments += tcDelta.Function.Arguments
					}

					// Stop if finished early without tool calls
					if choice.FinishReason != nil && *choice.FinishReason == "stop" {
						break
					}
				}
			}
			resp.Body.Close()

			for i := 0; i < len(toolCallsMap); i++ {
				if tc, ok := toolCallsMap[i]; ok {
					pendingToolCalls = append(pendingToolCalls, *tc)
				}
			}

			// If no tool calls, we are done
			if len(pendingToolCalls) == 0 {
				// Parse the assistantMessageContent to extract severity and code block for the UI
				severity := "WARNING"
				upperContent := strings.ToUpper(assistantMessageContent)
				if strings.Contains(upperContent, "BREAKING") {
					severity = "BREAKING"
				} else if strings.Contains(upperContent, "SAFE") {
					severity = "SAFE"
				}

				var code, language string
				startIdx := strings.Index(assistantMessageContent, "```")
				if startIdx != -1 {
					endIdx := strings.Index(assistantMessageContent[startIdx+3:], "```")
					if endIdx != -1 {
						codeBlock := assistantMessageContent[startIdx+3 : startIdx+3+endIdx]
						lines := strings.SplitN(codeBlock, "\n", 2)
						if len(lines) == 2 {
							language = strings.TrimSpace(lines[0])
							code = strings.TrimSpace(lines[1])
						} else {
							code = strings.TrimSpace(codeBlock)
						}
					}
				}

				writeSSE(w, flusher, SSEEvent{Type: "finding", Severity: severity, Content: "AI Analysis Complete (See above details)."})
				if code != "" {
					writeSSE(w, flusher, SSEEvent{Type: "fix", Language: language, Code: code})
				}

				break
			}

			// Execute tool calls
			assistantMsg := chatMessage{
				Role:      "assistant",
				Content:   assistantMessageContent,
				ToolCalls: pendingToolCalls,
			}
			messages = append(messages, assistantMsg)

			for _, tc := range pendingToolCalls {
				writeSSE(w, flusher, SSEEvent{Type: "thinking", Content: fmt.Sprintf("\n* Executing tool: %s *\n", tc.Function.Name)})

				result, err := executeTool(tc.Function.Name, tc.Function.Arguments, req)
				if err != nil {
					result = fmt.Sprintf("Error: %v", err)
				}

				messages = append(messages, chatMessage{
					Role:       "tool",
					Content:    result,
					Name:       tc.Function.Name,
					ToolCallID: tc.ID,
				})
			}

			// Continue loop to send tool results back to LLM
		}

		writeSSE(w, flusher, SSEEvent{Type: "done"})
	}
}
