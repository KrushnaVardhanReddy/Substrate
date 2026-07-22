package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type BreakingChange struct {
	RuleID      string `json:"rule_id"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Severity    string `json:"severity"`
}

type AIAutofixRequest struct {
	ProviderRepo       string           `json:"provider_repo"`
	SchemaType         string           `json:"schema_type"` // "openapi" | "sql" | "graphql"
	CurrentSchema      string           `json:"current_schema"`
	ProposedSchema     string           `json:"proposed_schema"`
	BreakingChanges    []BreakingChange `json:"breaking_changes"`
	ConsumerSourceCode string           `json:"consumer_source_code,omitempty"`
}

type AIAutofixResponse struct {
	Explanation   string `json:"explanation"`    // 2-3 sentence plain-English summary
	SafePatch     string `json:"safe_patch"`     // The exact corrected schema snippet or unified diff
	PatchLanguage string `json:"patch_language"` // "yaml" | "sql" | "graphql" | "diff"
	MockMode      bool   `json:"mock_mode"`      // true if SUBSTRATE_AI_BASE_URL is not set
}

func GenerateAutofixPatch(req AIAutofixRequest) (AIAutofixResponse, error) {
	aiConfig := loadAIConfig()
	if aiConfig.BaseURL == "" {
		return mockAutofixResponse(req), nil
	}

	breakingChangesJSON, _ := json.Marshal(req.BreakingChanges)

	var autofixSystemPrompt string
	var userMessage string

	if req.ConsumerSourceCode != "" {
		autofixSystemPrompt = `You are Substrate AI, an expert software engineer.
You will be given a provider's schema change, detected breaking changes, and a downstream consumer's source code.
Your job is to:
1. Write a 2-3 sentence plain-English explanation of how the schema change impacts the consumer code.
2. Generate a unified diff patch for the downstream consumer code to fix the breaking change so it complies with the new upstream schema.
3. Return ONLY valid JSON matching this schema:
   {"explanation": "...", "safe_patch": "...", "patch_language": "diff"}
Do not include markdown fences or extra text outside the JSON.`

		userMessage = fmt.Sprintf("Schema type: %s\nRepository: %s\n\nCurrent schema:\n%s\n\nProposed schema:\n%s\n\nDetected breaking changes:\n%s\n\nConsumer Source Code:\n%s",
			req.SchemaType, req.ProviderRepo, req.CurrentSchema, req.ProposedSchema, string(breakingChangesJSON), req.ConsumerSourceCode)
	} else {
		autofixSystemPrompt = `You are Substrate AI, an API schema safety expert.
You will be given a provider's current and proposed schema, plus a list of breaking changes detected.
Your job is to:
1. Write a 2-3 sentence plain-English explanation of the impact on downstream consumers.
2. Generate the minimal safe schema patch the provider should apply instead.
   - For deleted fields: restore them with a @deprecated marker.
   - For removed endpoints: add a deprecation notice instead of removing.
   - For type changes: keep the old field and add a new one.
3. Return ONLY valid JSON matching this schema:
   {"explanation": "...", "safe_patch": "...", "patch_language": "yaml|sql|graphql"}
Do not include markdown fences or extra text outside the JSON.`

		userMessage = fmt.Sprintf("Schema type: %s\nRepository: %s\n\nCurrent schema:\n%s\n\nProposed schema:\n%s\n\nDetected breaking changes:\n%s",
			req.SchemaType, req.ProviderRepo, req.CurrentSchema, req.ProposedSchema, string(breakingChangesJSON))
	}

	payload := map[string]interface{}{
		"model":           aiConfig.Model,
		"stream":          false,
		"response_format": map[string]string{"type": "json_object"},
		"messages": []map[string]string{
			{"role": "system", "content": autofixSystemPrompt},
			{"role": "user", "content": userMessage},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return AIAutofixResponse{}, fmt.Errorf("failed to marshal payload: %v", err)
	}

	reqObj, err := http.NewRequest(http.MethodPost, aiConfig.BaseURL+"/chat/completions", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return AIAutofixResponse{}, fmt.Errorf("failed to create request: %v", err)
	}

	reqObj.Header.Set("Authorization", "Bearer "+aiConfig.APIKey)
	reqObj.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(reqObj)
	if err != nil {
		return AIAutofixResponse{}, fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return AIAutofixResponse{}, fmt.Errorf("LLM returned status %d: %s", resp.StatusCode, string(body))
	}

	var llmResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return AIAutofixResponse{}, fmt.Errorf("failed to decode LLM response: %v", err)
	}

	if len(llmResp.Choices) == 0 {
		return AIAutofixResponse{}, fmt.Errorf("LLM returned no choices")
	}

	content := llmResp.Choices[0].Message.Content
	// Optional: trim markdown fences if the LLM still returns them despite the prompt
	content = strings.TrimPrefix(content, "```json\n")
	content = strings.TrimSuffix(content, "\n```")
	content = strings.TrimSuffix(content, "```")

	var autofixResp AIAutofixResponse
	if err := json.Unmarshal([]byte(content), &autofixResp); err != nil {
		return AIAutofixResponse{}, fmt.Errorf("failed to unmarshal LLM content: %v", err)
	}

	autofixResp.MockMode = false
	return autofixResp, nil
}

func AIAutofixHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var req AIAutofixRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		resp, err := GenerateAutofixPatch(req)
		if err != nil {
			writeError(w, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func mockAutofixResponse(req AIAutofixRequest) AIAutofixResponse {
	if req.ConsumerSourceCode != "" {
		return AIAutofixResponse{
			Explanation:   "Mock mode: Removing 'user_id' breaks the downstream consumer. We generated a unified diff to replace it with 'account_id'. Set SUBSTRATE_AI_BASE_URL to enable real AI analysis.",
			SafePatch:     "--- consumer.go\n+++ consumer.go\n@@ -10,3 +10,3 @@\n- userID := req.user_id\n+ userID := req.account_id",
			PatchLanguage: "diff",
			MockMode:      true,
		}
	}

	return AIAutofixResponse{
		Explanation:   "Mock mode: Removing 'user_id' from the response schema will break downstream consumers that read this field. Set SUBSTRATE_AI_BASE_URL to enable real AI analysis.",
		SafePatch:     "# Restore the field with a deprecation marker:\nuser_id:\n  type: string\n  deprecated: true\n  description: \"Deprecated — use account_id instead.\"",
		PatchLanguage: "yaml",
		MockMode:      true,
	}
}

func writeError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{"error": "AI autofix failed: " + msg})
}
