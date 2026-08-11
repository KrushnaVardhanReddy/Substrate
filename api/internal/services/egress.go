// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package services

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

type EventData struct {
	Organization    string           `json:"organization"`
	ProviderRepo    string           `json:"provider_repo"`
	PRNumber        int              `json:"pr_number,omitempty"`
	CommitSHA       string           `json:"commit_sha"`
	BreakingCount   int              `json:"breaking_count"`
	BrokenConsumers []BrokenConsumer `json:"broken_consumers"`
	DiffURL         string           `json:"diff_url"`
}

type BrokenConsumer struct {
	Repo       string   `json:"repo"`
	Codeowners []string `json:"codeowners"`
}

type BreakingChangeEvent struct {
	EventID   string    `json:"event_id"`
	EventType string    `json:"event_type"` // e.g., "substrate.breaking_change.detected"
	Timestamp string    `json:"timestamp"`
	Data      EventData `json:"data"`
}

// DispatchWebhook attempts to send a webhook event to a specific URL with optional secret signing.
// This executes the logic synchronously for a single webhook destination.
func DispatchWebhook(ctx context.Context, hook db.WebhookConfig, event BreakingChangeEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("error marshaling event: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}

	req, err := http.NewRequestWithContext(ctx, "POST", hook.URL, bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("error creating request for webhook %s: %w", hook.URL, err)
	}

	req.Header.Set("Content-Type", "application/json")

	if hook.Secret != "" {
		mac := hmac.New(sha256.New, []byte(hook.Secret))
		mac.Write(payload)
		signature := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Hub-Signature-256", "sha256="+signature)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("received non-2xx status code: %d", resp.StatusCode)
	}

	return nil
}

// We will also add a DispatchEvent which fetches webhooks, but returns them
// so a worker or the caller can enqueue a separate job per webhook config
// or handle it in one batch.
func GetWebhooksForOrg(ctx context.Context, store db.Store, org string) ([]db.WebhookConfig, error) {
	webhooks, err := store.GetWebhooks(ctx, org)
	if err != nil {
		log.Printf("Error fetching webhooks for org %s: %v", org, err)
		return nil, err
	}
	return webhooks, nil
}
