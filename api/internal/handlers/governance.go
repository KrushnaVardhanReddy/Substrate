// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/config"
	"github.com/google/cel-go/cel"
	"github.com/sashabaranov/go-openai"
)

type GenerateCELRequest struct {
	Prompt string `json:"prompt"`
}

type GenerateCELResponse struct {
	CEL   string `json:"cel,omitempty"`
	Error string `json:"error,omitempty"`
}

// GenerateCELHandler translates natural language to a CEL expression.
func GenerateCELHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		var req GenerateCELRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(GenerateCELResponse{Error: "invalid request body"})
			return
		}

		if strings.TrimSpace(req.Prompt) == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(GenerateCELResponse{Error: "prompt is required"})
			return
		}

		llmCfg := config.LoadLLMConfig()

		var celStr string
		if llmCfg.APIKey == "" {
			// Deterministic fallback mock response when AI keys are unset
			celStr = `request.path.matches('^/api/payment/') && !request.headers.contains('authorization')`
		} else {
			clientConfig := openai.DefaultConfig(llmCfg.APIKey)
			if llmCfg.BaseURL != "" {
				clientConfig.BaseURL = llmCfg.BaseURL
			}

			client := openai.NewClientWithConfig(clientConfig)

			ctx, cancel := context.WithTimeout(r.Context(), 1*time.Minute)
			defer cancel()

			systemPrompt := `You are an expert system that translates natural language governance rules into Common Expression Language (CEL) syntax.
Return ONLY the raw CEL expression. Do not include markdown formatting, backticks, or explanations.`

			resp, err := client.CreateChatCompletion(
				ctx,
				openai.ChatCompletionRequest{
					Model: llmCfg.Model,
					Messages: []openai.ChatCompletionMessage{
						{
							Role:    openai.ChatMessageRoleSystem,
							Content: systemPrompt,
						},
						{
							Role:    openai.ChatMessageRoleUser,
							Content: req.Prompt,
						},
					},
				},
			)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(GenerateCELResponse{Error: fmt.Sprintf("AI provider error: %v", err)})
				return
			}

			if len(resp.Choices) > 0 {
				celStr = strings.TrimSpace(resp.Choices[0].Message.Content)
				// Clean up any potential markdown ticks if the AI didn't follow instructions perfectly
				celStr = strings.TrimPrefix(celStr, "```cel")
				celStr = strings.TrimPrefix(celStr, "```")
				celStr = strings.TrimSuffix(celStr, "```")
				celStr = strings.TrimSpace(celStr)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(GenerateCELResponse{Error: "no response from AI provider"})
				return
			}
		}

		// Validate the CEL expression
		env, err := cel.NewEnv(
			cel.Variable("request.path", cel.StringType),
			cel.Variable("request.headers", cel.MapType(cel.StringType, cel.StringType)),
			cel.Variable("request.method", cel.StringType),
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(GenerateCELResponse{Error: fmt.Sprintf("failed to create CEL env: %v", err)})
			return
		}

		_, iss := env.Parse(celStr)
		if iss.Err() != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(GenerateCELResponse{Error: fmt.Sprintf("invalid CEL expression generated: %v", iss.Err())})
			return
		}

		json.NewEncoder(w).Encode(GenerateCELResponse{CEL: celStr})
	}
}
