// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package integrations

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
)

type KongClient struct {
	adminURL string
	token    string
	client   *http.Client
}

func NewKongClient(adminURL, token string) *KongClient {
	return &KongClient{
		adminURL: adminURL,
		token:    token,
		client:   &http.Client{},
	}
}

// PushServiceSpec uploads an API specification to Kong
func (c *KongClient) PushServiceSpec(ctx context.Context, serviceID string, spec []byte) error {
	url := fmt.Sprintf("%s/services/%s/document", c.adminURL, serviceID)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewBuffer(spec))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Kong-Admin-Token", c.token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
