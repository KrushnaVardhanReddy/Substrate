// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package postmortem

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/pkg/ai"
	"github.com/sashabaranov/go-openai"
)

type BreakingChangeRecord struct {
	OrgName         string          `json:"org_name"`
	RepoName        string          `json:"repo_name"`
	GitSHA          string          `json:"git_sha"`
	Timestamp       time.Time       `json:"timestamp"`
	BreakingChanges json.RawMessage `json:"breaking_changes"`
}

// GeneratePostMortem fetches breaking changes 24 hours prior to incidentDate and uses the AI client to generate a Markdown blameless post-mortem.
func GeneratePostMortem(ctx context.Context, apiURL, token string, incidentDate time.Time, out io.Writer) error {
	since := incidentDate.Add(-24 * time.Hour)
	q := url.Values{}
	q.Add("since", since.Format(time.RFC3339))
	q.Add("until", incidentDate.Format(time.RFC3339))
	reqURL := fmt.Sprintf("%s/api/v1/changes?%s", apiURL, q.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch breaking changes: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var changes []BreakingChangeRecord
	if err := json.NewDecoder(resp.Body).Decode(&changes); err != nil {
		return fmt.Errorf("failed to decode changes: %w", err)
	}

	if len(changes) == 0 {
		fmt.Fprintln(out, "No breaking changes found in the 24 hours prior to the incident.")
		return nil
	}

	// Format changes for the prompt
	var promptBuilder strings.Builder
	promptBuilder.WriteString(fmt.Sprintf("Incident Date: %s\n\n", incidentDate.Format(time.RFC3339)))
	promptBuilder.WriteString("Breaking Changes detected in the prior 24 hours:\n")
	for _, change := range changes {
		promptBuilder.WriteString(fmt.Sprintf("- Repository: %s/%s\n  Commit: %s\n  Timestamp: %s\n  Changes: %s\n\n",
			change.OrgName, change.RepoName, change.GitSHA, change.Timestamp.Format(time.RFC3339), string(change.BreakingChanges)))
	}

	aiClient, err := ai.NewAIClient()
	if err != nil {
		return fmt.Errorf("failed to initialize AI client: %w", err)
	}

	systemPrompt := "You are an expert SRE and Technical Writer. Based on the provided database of breaking schema changes that occurred shortly before an incident, draft a Blameless Post-Mortem document. Include the following sections:\n1. Summary (Cost & Impact Estimate)\n2. Root Cause (The breaking schema change)\n3. Blast Radius (Downstream consumers affected)\n4. Remediation & Action Items\n\nOutput MUST be fully formatted in Markdown and ready to share."

	chatReq := openai.ChatCompletionRequest{
		Model: openai.GPT4o,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: promptBuilder.String(),
			},
		},
		Stream: true,
	}

	stream, err := aiClient.CreateChatCompletionStream(ctx, chatReq)
	if err != nil {
		return fmt.Errorf("failed to create chat completion stream: %w", err)
	}
	defer stream.Close()

	fmt.Fprintln(out, "Generating Post-Mortem...")

	for {
		chunk, err := stream.Recv()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			// Handle EOF wrapped or normal
			if strings.Contains(err.Error(), "EOF") {
				break
			}
			return fmt.Errorf("stream error: %w", err)
		}
		fmt.Fprint(out, chunk)
	}

	fmt.Fprintln(out)
	return nil
}
