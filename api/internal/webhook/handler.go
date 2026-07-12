package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/KrushnaVardhanReddy/Substrate/api/internal/integrations/postman"
)

type WebhookPayload struct {
	Event   string `json:"event"`
	Status  string `json:"status"` // "SAFE", "APPROVED", etc.
	Schema  string `json:"schema"` // The OpenAPI schema JSON/YAML string
	Type    string `json:"type"`   // "openapi", etc.
	Version string `json:"version"`
}

type SyncClient interface {
	SyncSchema(ctx context.Context, schema string) error
}

func Handler() http.HandlerFunc {
	return HandlerWithClient(postman.NewClient())
}

func HandlerWithClient(client SyncClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload WebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid payload", http.StatusBadRequest)
			return
		}

		if payload.Type == "openapi" && (payload.Status == "SAFE" || payload.Status == "APPROVED") {
			ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
			defer cancel()
			err := client.SyncSchema(ctx, payload.Schema)
			if err != nil {
				http.Error(w, "failed to sync to postman", http.StatusInternalServerError)
				return
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
