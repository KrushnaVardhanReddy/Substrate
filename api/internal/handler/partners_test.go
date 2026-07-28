package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
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

	handler := NewPartnersHandler(db.NewPGStore(pool))
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

	handler := NewPartnersHandler(db.NewPGStore(pool))

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

	handler := NewPartnersHandler(db.NewPGStore(pool))

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
func TestPartnersHandler_UpdatePartner(t *testing.T) {
	ctx := context.Background()
	pool, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	// Seed data
	partnerID := "33333333-3333-3333-3333-333333333333"
	_, err := pool.Exec(ctx, `
		INSERT INTO partner_integrations (id, vendor_name, webhook_url, webhook_secret, status)
		VALUES ($1, 'Old Vendor', 'http://old.com', 'old-secret', 'pending')
	`, partnerID)
	require.NoError(t, err)

	handler := NewPartnersHandler(db.NewPGStore(pool))

	tests := []struct {
		name           string
		partnerID      string
		payload        map[string]interface{}
		expectedStatus int
		expectedName   string
	}{
		{
			name:      "Update Success",
			partnerID: partnerID,
			payload: map[string]interface{}{
				"vendor_name": "Updated Vendor",
				"webhook_url": "http://new.com",
			},
			expectedStatus: http.StatusOK,
			expectedName:   "Updated Vendor",
		},
		{
			name:      "Not Found",
			partnerID: "99999999-9999-9999-9999-999999999999",
			payload: map[string]interface{}{
				"vendor_name": "Ghost Vendor",
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req := httptest.NewRequest("PUT", "/partners/"+tt.partnerID, bytes.NewBuffer(body))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.partnerID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.UpdatePartner(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var p PartnerIntegration
				json.Unmarshal(rr.Body.Bytes(), &p)
				assert.Equal(t, tt.expectedName, p.VendorName)
			}
		})
	}
}

func TestPartnersHandler_DeletePartner(t *testing.T) {
	ctx := context.Background()
	pool, cleanup := testutils.SetupTestDB(t)
	defer cleanup()

	// Seed data
	partnerID := "44444444-4444-4444-4444-444444444444"
	_, err := pool.Exec(ctx, `
		INSERT INTO partner_integrations (id, vendor_name, webhook_url, webhook_secret, status)
		VALUES ($1, 'Delete Vendor', 'http://del.com', 'del-secret', 'pending')
	`, partnerID)
	require.NoError(t, err)

	handler := NewPartnersHandler(db.NewPGStore(pool))

	tests := []struct {
		name           string
		partnerID      string
		expectedStatus int
	}{
		{
			name:           "Delete Success",
			partnerID:      partnerID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Not Found",
			partnerID:      "99999999-9999-9999-9999-999999999999",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("DELETE", "/partners/"+tt.partnerID, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.partnerID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.DeletePartner(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
