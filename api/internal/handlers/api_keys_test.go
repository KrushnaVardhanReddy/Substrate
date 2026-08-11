// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/handlers"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
)

type mockAPIKeyStore struct {
	ports.Store
	keys []*db.APIKey
}

func (m *mockAPIKeyStore) GetOrgIDByName(ctx context.Context, name string) (uuid.UUID, error) {
	return uuid.MustParse("00000000-0000-0000-0000-000000000001"), nil
}

func (m *mockAPIKeyStore) CreateAPIKey(ctx context.Context, orgID uuid.UUID, name, prefix, hash string) (*db.APIKey, error) {
	key := &db.APIKey{
		ID:        uuid.New(),
		OrgID:     orgID,
		Name:      name,
		Prefix:    prefix,
		Hash:      hash,
		CreatedAt: time.Now(),
	}
	m.keys = append(m.keys, key)
	return key, nil
}

func (m *mockAPIKeyStore) ListAPIKeys(ctx context.Context, orgID uuid.UUID) ([]*db.APIKey, error) {
	var result []*db.APIKey
	for _, k := range m.keys {
		if k.OrgID == orgID {
			result = append(result, k)
		}
	}
	return result, nil
}

func (m *mockAPIKeyStore) DeleteAPIKey(ctx context.Context, id, orgID uuid.UUID) error {
	var remaining []*db.APIKey
	for _, k := range m.keys {
		if k.ID != id || k.OrgID != orgID {
			remaining = append(remaining, k)
		}
	}
	m.keys = remaining
	return nil
}

func TestAPIKeyHandler(t *testing.T) {
	mockStore := &mockAPIKeyStore{}
	h := &handlers.APIKeyHandler{Store: mockStore}

	r := chi.NewRouter()
	r.Post("/api/v1/org/{org}/apikeys", h.CreateAPIKey)
	r.Get("/api/v1/org/{org}/apikeys", h.ListAPIKeys)
	r.Delete("/api/v1/org/{org}/apikeys/{id}", h.DeleteAPIKey)

	// Create
	reqBody := `{"name": "test-key"}`
	req, _ := http.NewRequest("POST", "/api/v1/org/test-org/apikeys", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var resp struct {
		Key      *db.APIKey `json:"key"`
		RawToken string     `json:"raw_token"`
	}
	json.NewDecoder(rr.Body).Decode(&resp)

	if resp.RawToken == "" {
		t.Errorf("expected raw_token in response")
	}
	if len(resp.RawToken) != 64 { // 32 bytes hex encoded
		t.Errorf("expected raw_token to be 64 chars, got %d", len(resp.RawToken))
	}
	if resp.Key.Name != "test-key" {
		t.Errorf("expected key name 'test-key', got '%s'", resp.Key.Name)
	}

	keyID := resp.Key.ID.String()

	// List
	req, _ = http.NewRequest("GET", "/api/v1/org/test-org/apikeys", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var listResp []*db.APIKey
	json.NewDecoder(rr.Body).Decode(&listResp)

	if len(listResp) != 1 {
		t.Errorf("expected 1 key in list, got %d", len(listResp))
	}

	// Delete
	req, _ = http.NewRequest("DELETE", "/api/v1/org/test-org/apikeys/"+keyID, nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNoContent)
	}

	// List again to verify deletion
	req, _ = http.NewRequest("GET", "/api/v1/org/test-org/apikeys", nil)
	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	json.NewDecoder(rr.Body).Decode(&listResp)
	if len(listResp) != 0 {
		t.Errorf("expected 0 keys in list, got %d", len(listResp))
	}
}
