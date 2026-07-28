package webhook

import (
	"encoding/json"
	"log"
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
		var rawPayload json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&rawPayload); err != nil {
			http.Error(w, `{"error": "invalid json"}`, http.StatusBadRequest)
			return
		}
		log.Printf("Received Webhook: %s", string(rawPayload))

		ctx := r.Context()
		var req services.PushPayload

		// Try to decode as Gitea Webhook first (check for "repository" and "commits")
		var giteaWebhook GiteaPushWebhook
		if err := json.Unmarshal(rawPayload, &giteaWebhook); err == nil && giteaWebhook.Repository.Name != "" && len(giteaWebhook.Commits) > 0 {
			req = services.PushPayload{
				InstallationID: 0,
				Org:            giteaWebhook.Repository.Owner.Login,
				Repo:           giteaWebhook.Repository.FullName,
				GithubRepoID:   giteaWebhook.Repository.ID,
				CommitSHA:      giteaWebhook.After,
				Files:          []services.File{},
			}
			
			// For each commit, collect added and modified files
			filesToFetch := make(map[string]bool)
			for _, commit := range giteaWebhook.Commits {
				for _, f := range commit.Added {
					filesToFetch[f] = true
				}
				for _, f := range commit.Modified {
					filesToFetch[f] = true
				}
			}
			
			for filePath := range filesToFetch {
				content, err := ghClient.GetFileContent(ctx, giteaWebhook.Repository.Owner.Login, giteaWebhook.Repository.Name, filePath)
				if err == nil {
					req.Files = append(req.Files, services.File{
						Path:    filePath,
						Content: content,
					})
				} else {
					log.Printf("Failed to get file content for %s: %v", filePath, err)
				}
			}
		} else {
			// Fallback to standard PushPayload
			if err := json.Unmarshal(rawPayload, &req); err != nil {
				http.Error(w, `{"error": "invalid json"}`, http.StatusBadRequest)
				return
			}
		}

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
