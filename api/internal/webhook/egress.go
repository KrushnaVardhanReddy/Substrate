package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

func DispatchEvent(ctx context.Context, store db.Store, event BreakingChangeEvent) error {
	webhooks, err := store.GetWebhooks(ctx, event.Data.Organization)
	if err != nil {
		log.Printf("Error fetching webhooks for org %s: %v", event.Data.Organization, err)
		return err
	}

	if len(webhooks) == 0 {
		return nil
	}

	payload, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling event: %v", err)
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, hook := range webhooks {
		go func(h db.WebhookConfig) {
			// Retry mechanism
			maxRetries := 3
			backoff := 1 * time.Second

			for i := 0; i < maxRetries; i++ {
				req, err := http.NewRequestWithContext(context.Background(), "POST", h.URL, bytes.NewBuffer(payload))
				if err != nil {
					log.Printf("Error creating request for webhook %s: %v", h.URL, err)
					return
				}

				req.Header.Set("Content-Type", "application/json")

				if h.Secret != "" {
					mac := hmac.New(sha256.New, []byte(h.Secret))
					mac.Write(payload)
					signature := hex.EncodeToString(mac.Sum(nil))
					req.Header.Set("X-Hub-Signature-256", "sha256="+signature)
				}

				resp, err := client.Do(req)
				if err == nil {
					resp.Body.Close()
					if resp.StatusCode >= 200 && resp.StatusCode < 300 {
						// Success
						return
					}
				}

				if i < maxRetries-1 {
					time.Sleep(backoff)
					backoff *= 2
				}
			}
			log.Printf("Failed to deliver webhook to %s after %d retries", h.URL, maxRetries)
		}(hook)
	}

	return nil
}
