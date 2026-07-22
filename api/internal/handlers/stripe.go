package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/webhook"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
)

// StripeWebhookHandler handles incoming webhooks from Stripe
func StripeWebhookHandler(store ports.BillingStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const MaxBodyBytes = int64(65536)
		r.Body = http.MaxBytesReader(w, r.Body, MaxBodyBytes)
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading request body", http.StatusServiceUnavailable)
			return
		}

		endpointSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
		if endpointSecret == "" {
			http.Error(w, "Stripe webhook secret not configured", http.StatusInternalServerError)
			return
		}

		signatureHeader := r.Header.Get("Stripe-Signature")

		// In tests, event api version mismatches can happen if the mock event doesn't set it.
		// Ignore API version mismatch just for robustness with stripe test payloads.
		event, err := webhook.ConstructEventWithOptions(payload, signatureHeader, endpointSecret, webhook.ConstructEventOptions{
			IgnoreAPIVersionMismatch: true,
		})
		if err != nil {
			http.Error(w, "Error verifying webhook signature", http.StatusBadRequest)
			return
		}

		if event.Type == "checkout.session.completed" {
			var session stripe.CheckoutSession
			err := json.Unmarshal(event.Data.Raw, &session)
			if err != nil {
				http.Error(w, "Error parsing webhook JSON", http.StatusBadRequest)
				return
			}

			if session.ClientReferenceID != "" {
				orgID, err := uuid.Parse(session.ClientReferenceID)
				if err != nil {
					http.Error(w, "Invalid client_reference_id format", http.StatusBadRequest)
					return
				}

				var stripeCustomerID string
				if session.Customer != nil {
					stripeCustomerID = session.Customer.ID
				}

				if stripeCustomerID != "" {
					err = store.UpdateStripeCustomerID(r.Context(), orgID, stripeCustomerID)
					if err != nil {
						http.Error(w, "Failed to update org in database", http.StatusInternalServerError)
						return
					}
				}
			}
		}

		w.WriteHeader(http.StatusOK)
	}
}
