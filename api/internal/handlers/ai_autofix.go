package handlers

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
	ProviderRepo    string           `json:"provider_repo"`
	SchemaType      string           `json:"schema_type"` // "openapi" | "sql" | "graphql"
	CurrentSchema   string           `json:"current_schema"`
	ProposedSchema  string           `json:"proposed_schema"`
	BreakingChanges []BreakingChange `json:"breaking_changes"`
}

type AIAutofixResponse struct {
	Explanation   string `json:"explanation"`    // 2-3 sentence plain-English summary
	SafePatch     string `json:"safe_patch"`     // The exact corrected schema snippet
	PatchLanguage string `json:"patch_language"` // "yaml" | "sql" | "graphql"
	MockMode      bool   `json:"mock_mode"`      // true if SUBSTRATE_AI_BASE_URL is not set
}

const autofixSystemPrompt = `You are Substrate AI, an API schema safety expert.
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

		aiConfig := loadAIConfig()
		if aiConfig.BaseURL == "" {
			mockAutofixResponse(w)
			return
		}

		breakingChangesJSON, _ := json.Marshal(req.BreakingChanges)
		userMessage := fmt.Sprintf("Schema type: %s\nRepository: %s\n\nCurrent schema:\n%s\n\nProposed schema:\n%s\n\nDetected breaking changes:\n%s",
			req.SchemaType, req.ProviderRepo, req.CurrentSchema, req.ProposedSchema, string(breakingChangesJSON))

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
			writeError(w, fmt.Sprintf("failed to marshal payload: %v", err))
			return
		}

		reqObj, err := http.NewRequest(http.MethodPost, aiConfig.BaseURL+"/chat/completions", bytes.NewBuffer(payloadBytes))
		if err != nil {
			writeError(w, fmt.Sprintf("failed to create request: %v", err))
			return
		}

		reqObj.Header.Set("Authorization", "Bearer "+aiConfig.APIKey)
		reqObj.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(reqObj)
		if err != nil {
			writeError(w, fmt.Sprintf("failed to execute request: %v", err))
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			writeError(w, fmt.Sprintf("LLM returned status %d: %s", resp.StatusCode, string(body)))
			return
		}

		var llmResp struct {
			Choices []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			} `json:"choices"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
			writeError(w, fmt.Sprintf("failed to decode LLM response: %v", err))
			return
		}

		if len(llmResp.Choices) == 0 {
			writeError(w, "LLM returned no choices")
			return
		}

		content := llmResp.Choices[0].Message.Content
		// Optional: trim markdown fences if the LLM still returns them despite the prompt
		content = strings.TrimPrefix(content, "```json\n")
		content = strings.TrimSuffix(content, "\n```")
		content = strings.TrimSuffix(content, "```")

		var autofixResp AIAutofixResponse
		if err := json.Unmarshal([]byte(content), &autofixResp); err != nil {
			writeError(w, fmt.Sprintf("failed to unmarshal LLM content: %v", err))
			return
		}

		autofixResp.MockMode = false

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(autofixResp)
	}
}

func mockAutofixResponse(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AIAutofixResponse{
		Explanation:   "Mock mode: Removing 'user_id' from the response schema will break downstream consumers that read this field. Set SUBSTRATE_AI_BASE_URL to enable real AI analysis.",
		SafePatch:     "# Restore the field with a deprecation marker:\nuser_id:\n  type: string\n  deprecated: true\n  description: \"Deprecated — use account_id instead.\"",
		PatchLanguage: "yaml",
		MockMode:      true,
	})
}

func writeError(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(map[string]string{"error": "AI autofix failed: " + msg})
}
