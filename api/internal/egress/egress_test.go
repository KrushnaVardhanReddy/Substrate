package egress

import (
	"context"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/workers"
)

func TestDispatchEvent(t *testing.T) {
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
				{Org: "acme-corp", URL: "http://example.com", Secret: "my-secret"},
			},
			expectCalled: true,
			secret:       "my-secret",
		},
		{
			name: "Success without secret",
			org:  "acme-corp",
			webhooks: []db.WebhookConfig{
				{Org: "acme-corp", URL: "http://example.com"},
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

	// We can create an actual pgxpool for testing if needed, or simply pass nil and recover the panic in test
	// But it's easier to just pass a mock or accept it will panic if nil, let's use recover in the test since we just want to ensure it doesn't fail on db mock.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

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

			mockEnqueuer := &workers.MockJobEnqueuer{}

			err := DispatchEvent(context.Background(), mockStore, mockEnqueuer, event)
			if err != nil {
				// The only error returned directly is from GetWebhooks (mocked above)
				t.Fatalf("expected no direct error from GetWebhooks: %v", err)
			}
		})
	}
}
