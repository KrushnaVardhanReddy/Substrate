package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// A simple client just for E2E bypassing the internal package restriction
func CreateDraftPR(ctx context.Context, owner, repo, branch, patch, title, body string) (string, error) {
	apiURL := os.Getenv("GITHUB_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:3000/api/v1"
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = "mock-token"
	}
	prURL := fmt.Sprintf("%s/repos/%s/%s/pulls", apiURL, owner, repo)
	prPayload := map[string]interface{}{
		"title": title,
		"body":  body + "\n\n```diff\n" + patch + "\n```",
		"head":  branch,
		"base":  "main",
		"draft": true,
	}

	prBytes, _ := json.Marshal(prPayload)
	reqPR, err := http.NewRequestWithContext(ctx, http.MethodPost, prURL, bytes.NewBuffer(prBytes))
	if err != nil {
		return "", err
	}
	reqPR.Header.Set("Authorization", "token "+token)

	client := &http.Client{}
	respPR, err := client.Do(reqPR)
	if err != nil {
		return "", err
	}
	defer respPR.Body.Close()

	if respPR.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("failed: %d", respPR.StatusCode)
	}

	var result map[string]interface{}
	json.NewDecoder(respPR.Body).Decode(&result)
	u, _ := result["html_url"].(string)
	return u, nil
}

func ListIssueComments(ctx context.Context, owner, repo string, issueNumber int) ([]string, error) {
	apiURL := os.Getenv("GITHUB_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:3000/api/v1"
	}
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		token = "mock-token"
	}
	url := fmt.Sprintf("%s/repos/%s/%s/issues/%d/comments", apiURL, owner, repo, issueNumber)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed: %d", resp.StatusCode)
	}

	var comments []struct {
		Body string `json:"body"`
	}
	json.NewDecoder(resp.Body).Decode(&comments)

	var res []string
	for _, c := range comments {
		res = append(res, c.Body)
	}
	return res, nil
}
