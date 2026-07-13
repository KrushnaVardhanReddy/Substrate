package webhook

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

func TestDispatchEvent(t *testing.T) {
	// Create mock HTTP server to receive webhook
	requestsReceived := 0
	var lastReqBody string
	var lastReqSig string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestsReceived++
		body, _ := io.ReadAll(r.Body)
		lastReqBody = string(body)
		lastReqSig = r.Header.Get("X-Hub-Signature-256")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	tests := []struct {
		name         string
		org          string
		webhooks     []db.WebhookConfig
		expectCalled bool
		expectErr    bool
		secret       string
	}{
		{
			name: "Success with secret",
			org:  "acme-corp",
			webhooks: []db.WebhookConfig{
				{Org: "acme-corp", URL: server.URL, Secret: "my-secret"},
			},
			expectCalled: true,
			secret:       "my-secret",
		},
		{
			name: "Success without secret",
			org:  "acme-corp",
			webhooks: []db.WebhookConfig{
				{Org: "acme-corp", URL: server.URL},
			},
			expectCalled: true,
		},
		{
			name:         "No webhooks",
			org:          "acme-corp",
			webhooks:     []db.WebhookConfig{},
			expectCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestsReceived = 0

			mockStore := &db.MockStore{
				GetWebhooksFunc: func(ctx context.Context, org string) ([]db.WebhookConfig, error) {
					return tt.webhooks, nil
				},
			}

			event := BreakingChangeEvent{
				EventID:   "evt_1",
				EventType: "substrate.breaking_change.detected",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Data: EventData{
					Organization:  tt.org,
					ProviderRepo:  "acme-corp/api",
					CommitSHA:     "abc1234",
					BreakingCount: 1,
				},
			}

			err := DispatchEvent(context.Background(), mockStore, event)
			if (err != nil) != tt.expectErr {
				t.Fatalf("expected error: %v, got: %v", tt.expectErr, err)
			}

			// Since DispatchEvent uses goroutines, wait briefly
			time.Sleep(50 * time.Millisecond)

			if tt.expectCalled && requestsReceived == 0 {
				t.Errorf("expected webhook to be called, but it wasn't")
			} else if !tt.expectCalled && requestsReceived > 0 {
				t.Errorf("expected no webhook calls, but got %d", requestsReceived)
			}

			if tt.expectCalled {
				if !strings.Contains(lastReqBody, "abc1234") {
					t.Errorf("expected body to contain payload, got: %s", lastReqBody)
				}
				if tt.secret != "" && lastReqSig == "" {
					t.Errorf("expected signature header to be set")
				} else if tt.secret == "" && lastReqSig != "" {
					t.Errorf("expected no signature header, got: %s", lastReqSig)
				}
			}
		})
	}
}
