package webhook

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
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

		trialEndsAt, stripeCustomerID, err := store.GetBillingStatus(ctx, req.Org)
		if err == nil {
			if time.Now().After(trialEndsAt) && (stripeCustomerID == nil || *stripeCustomerID == "") {
				// Paywall Pause: trial expired and no stripe customer id
				err = ghClient.CreateCheckRun(
					ctx,
					req.Org,
					req.Repo,
					req.CommitSHA,
					"substrate",
					"Substrate Trial Expired",
					"Substrate 90-day trial has expired. Please visit the dashboard to upgrade to the Enterprise plan and resume API protection.",
					"failure",
				)
				if err != nil {
					// Log the error but continue returning the paused response
					// In a real system, you would log to an external service or stdout
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusAccepted)
				json.NewEncoder(w).Encode(map[string]string{"status": "paused_due_to_billing"})
				return
			}
		}

		_, err = riverClient.Insert(ctx, workers.PushWebhookJob{Payload: req}, nil)
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
