// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package github

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (c *RESTClient) RequestReviewers(ctx context.Context, owner, repo string, pullNumber int, reviewers []string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/requested_reviewers", c.apiURL, owner, repo, pullNumber)

	var users []string
	var teams []string

	for _, r := range reviewers {
		r = strings.TrimPrefix(r, "@")
		if strings.Contains(r, "/") {
			parts := strings.Split(r, "/")
			if len(parts) == 2 {
				teams = append(teams, parts[1])
			} else {
				teams = append(teams, r)
			}
		} else {
			users = append(users, r)
		}
	}

	payload := map[string]interface{}{
		"reviewers":      users,
		"team_reviewers": teams,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return err
	}
	c.addHeaders(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("request reviewers failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}
