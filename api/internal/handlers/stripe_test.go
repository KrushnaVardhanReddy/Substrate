// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/webhook"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

func TestStripeWebhookHandler(t *testing.T) {
	os.Setenv("STRIPE_WEBHOOK_SECRET", "whsec_test_secret")

	orgID := uuid.New()
	session := stripe.CheckoutSession{
		ClientReferenceID: orgID.String(),
		Customer: &stripe.Customer{
			ID: "cus_12345",
		},
	}
	sessionBytes, _ := json.Marshal(session)
	event := stripe.Event{
		Type: "checkout.session.completed",
		Data: &stripe.EventData{
			Raw: sessionBytes,
		},
	}

	t.Run("Invalid signature", func(t *testing.T) {
		mockStore := &db.MockStore{}
		handler := StripeWebhookHandler(mockStore)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/stripe", bytes.NewBuffer([]byte(`{}`)))
		req.Header.Set("Stripe-Signature", "t=123,v1=invalid")
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 Bad Request, got %d", rec.Code)
		}
	})

	t.Run("Valid checkout.session.completed", func(t *testing.T) {
		var updatedOrgID uuid.UUID
		var updatedStripeID string

		mockStore := &db.MockStore{
			UpdateStripeCustomerIDFunc: func(ctx context.Context, oID uuid.UUID, sID string) error {
				updatedOrgID = oID
				updatedStripeID = sID
				return nil
			},
		}

		handler := StripeWebhookHandler(mockStore)

		payloadBytes, _ := json.Marshal(event)

		sigPayload := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
			Payload:   payloadBytes,
			Secret:    "whsec_test_secret",
			Timestamp: time.Now(),
		})

		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/stripe", bytes.NewBuffer(payloadBytes))
		req.Header.Set("Stripe-Signature", sigPayload.Header)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Logf("Response body: %s", rec.Body.String())
			t.Errorf("Expected status 200 OK, got %d", rec.Code)
		}

		if updatedOrgID != orgID {
			t.Errorf("Expected updated org ID %s, got %s", orgID, updatedOrgID)
		}
		if updatedStripeID != "cus_12345" {
			t.Errorf("Expected updated stripe ID cus_12345, got %s", updatedStripeID)
		}
	})
}
