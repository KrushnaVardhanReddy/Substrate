package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/integrations/postman"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/services"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/workers"
)

// PushHandler handles the GitHub push webhook payload forwarded from the GitHub App
func PushHandler(store db.Store, ghClient github.Client, riverClient workers.JobEnqueuer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req services.PushPayload
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, `{"error": "invalid json"}`, http.StatusBadRequest)
			return
		}

		ctx := r.Context()

		_, err := riverClient.Insert(ctx, workers.PushWebhookJob{Payload: req}, nil)
		if err != nil {
			http.Error(w, `{"error": "internal error enqueuing push job"}`, http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "queued"})
	}
}

func extractProviderFromURL(urlStr string) string {
	urlStr = strings.TrimPrefix(urlStr, "http://")
	urlStr = strings.TrimPrefix(urlStr, "https://")

	parts := strings.Split(urlStr, ":")
	host := parts[0]

	hostParts := strings.Split(host, ".")
	if len(hostParts) > 0 {
		return hostParts[0]
	}
	return ""
}

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
