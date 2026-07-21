package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/testutils"
)

func TestPartnersHandler_ListPartners(t *testing.T) {
	ctx := context.Background()
	pool, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	// Seed data
	_, err := pool.Exec(ctx, `
		INSERT INTO partner_integrations (id, vendor_name, webhook_url, webhook_secret, status)
		VALUES ('11111111-1111-1111-1111-111111111111', 'Test Vendor', 'http://example.com', 'secret', 'pending')
	`)
	require.NoError(t, err)

	handler := NewPartnersHandler(pool)
	req := httptest.NewRequest("GET", "/partners", nil)
	rr := httptest.NewRecorder()

	handler.ListPartners(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var partners []PartnerIntegration
	err = json.Unmarshal(rr.Body.Bytes(), &partners)
	require.NoError(t, err)

	assert.Len(t, partners, 1)
	assert.Equal(t, "Test Vendor", partners[0].VendorName)
}

func TestPartnersHandler_CreatePartner(t *testing.T) {
	pool, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	handler := NewPartnersHandler(pool)

	tests := []struct {
		name           string
		payload        CreatePartnerRequest
		expectedStatus int
	}{
		{
			name: "Success",
			payload: CreatePartnerRequest{
				VendorName:    "New Vendor",
				WebhookURL:    "http://example.com/webhook",
				WebhookSecret: "super-secret",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Missing Fields",
			payload: CreatePartnerRequest{
				VendorName: "Missing URL",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("POST", "/partners", bytes.NewBuffer(body))
			rr := httptest.NewRecorder()

			handler.CreatePartner(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusCreated {
				var p PartnerIntegration
				json.Unmarshal(rr.Body.Bytes(), &p)
				assert.Equal(t, "New Vendor", p.VendorName)
				assert.Equal(t, "pending", p.Status)
			}
		})
	}
}

func TestPartnersHandler_VerifyPartner(t *testing.T) {
	ctx := context.Background()
	pool, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	// Mock webhook server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Assert signature exists
		if r.Header.Get("X-Substrate-Signature") == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	// Seed data
	partnerID := "22222222-2222-2222-2222-222222222222"
	_, err := pool.Exec(ctx, `
		INSERT INTO partner_integrations (id, vendor_name, webhook_url, webhook_secret, status)
		VALUES ($1, 'Verified Vendor', $2, 'test-secret', 'pending')
	`, partnerID, mockServer.URL)
	require.NoError(t, err)

	handler := NewPartnersHandler(pool)

	// Create request with chi routing context
	req := httptest.NewRequest("POST", "/partners/"+partnerID+"/verify", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", partnerID)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	handler.VerifyPartner(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var p PartnerIntegration
	err = json.Unmarshal(rr.Body.Bytes(), &p)
	require.NoError(t, err)
	assert.Equal(t, "certified", p.Status)

	// Verify in DB
	var status string
	err = pool.QueryRow(ctx, "SELECT status FROM partner_integrations WHERE id = $1", partnerID).Scan(&status)
	require.NoError(t, err)
	assert.Equal(t, "certified", status)
}
