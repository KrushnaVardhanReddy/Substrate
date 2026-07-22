package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type ChaosTestingService struct {
	aiConfig AIConfig
}

func NewChaosTestingService() *ChaosTestingService {
	return &ChaosTestingService{
		aiConfig: loadAIConfig(),
	}
}

func (s *ChaosTestingService) GenerateChaosTest(oldSchema, newSchema, breakingChanges string) (string, error) {
	if s.aiConfig.BaseURL == "" {
		return s.mockChaosTest(oldSchema, newSchema, breakingChanges), nil
	}

	systemPrompt := `You are Substrate AI Chaos Engineering Bot.
Your job is to write a single JavaScript test script (runnable in Node.js) that proves how a given schema change breaks an expected contract.
You will be provided with:
1. The old schema.
2. The new schema.
3. Detected breaking changes.

Requirements:
- Output ONLY valid JavaScript code. Do not include markdown fences, explanation, or extra text.
- The script should use "assert" module.
- It should mock a response based on the new schema and assert expectations based on the old schema.
- The script MUST throw an error or fail an assertion to prove the breakage.`

	userMessage := fmt.Sprintf("Old Schema:\n%s\n\nNew Schema:\n%s\n\nBreaking Changes:\n%s", oldSchema, newSchema, breakingChanges)

	payload := map[string]interface{}{
		"model":  s.aiConfig.Model,
		"stream": false,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userMessage},
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %v", err)
	}

	reqObj, err := http.NewRequest(http.MethodPost, s.aiConfig.BaseURL+"/chat/completions", bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	reqObj.Header.Set("Authorization", "Bearer "+s.aiConfig.APIKey)
	reqObj.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(reqObj)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("LLM returned status %d: %s", resp.StatusCode, string(body))
	}

	var llmResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&llmResp); err != nil {
		return "", fmt.Errorf("failed to decode LLM response: %v", err)
	}

	if len(llmResp.Choices) == 0 {
		return "", fmt.Errorf("LLM returned no choices")
	}

	content := llmResp.Choices[0].Message.Content
	content = strings.TrimPrefix(content, "```javascript\n")
	content = strings.TrimPrefix(content, "```js\n")
	content = strings.TrimSuffix(content, "\n```")
	content = strings.TrimSuffix(content, "```")

	return content, nil
}

func (s *ChaosTestingService) mockChaosTest(oldSchema, newSchema, breakingChanges string) string {
	return `const assert = require('assert');

// Mock response representing the NEW broken schema
const newResponse = {
	account_id: "acc_123",
	name: "Test User"
};

try {
	// The consumer expects the old schema (which had user_id)
	assert.ok(newResponse.user_id, "Expected user_id to be present in response");
} catch (err) {
	console.error("Chaos test proven breakage:", err.message);
	process.exit(1);
}
`
}
