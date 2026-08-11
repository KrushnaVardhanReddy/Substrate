// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package marketplace

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/google/uuid"
)

func TestPublishHandler(t *testing.T) {
	mockStore := &db.MockStore{
		PublishPluginFunc: func(ctx context.Context, name, description string, schemaContent json.RawMessage) (uuid.UUID, error) {
			return uuid.MustParse("00000000-0000-0000-0000-000000000001"), nil
		},
	}

	reqBody := `{"name": "test-plugin", "description": "test", "schema_content": {"rules": []}}`
	req, _ := http.NewRequest("POST", "/api/marketplace/publish", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := PublishHandler(mockStore)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected %d, got %d", http.StatusCreated, rr.Code)
	}
}

func TestListPluginsHandler(t *testing.T) {
	mockStore := &db.MockStore{
		ListPluginsFunc: func(ctx context.Context) ([]db.MarketplacePlugin, error) {
			return []db.MarketplacePlugin{
				{
					ID:            uuid.MustParse("00000000-0000-0000-0000-000000000001"),
					Name:          "test-plugin",
					Description:   "test",
					SchemaContent: json.RawMessage(`{"rules": []}`),
					CreatedAt:     time.Now(),
					UpdatedAt:     time.Now(),
				},
			}, nil
		},
	}

	req, _ := http.NewRequest("GET", "/api/marketplace/plugins", nil)

	rr := httptest.NewRecorder()
	handler := ListPluginsHandler(mockStore)

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected %d, got %d", http.StatusOK, rr.Code)
	}

	var plugins []db.MarketplacePlugin
	if err := json.NewDecoder(rr.Body).Decode(&plugins); err != nil {
		t.Fatal(err)
	}

	if len(plugins) != 1 || plugins[0].Name != "test-plugin" {
		t.Errorf("expected 1 plugin with name 'test-plugin', got %v", plugins)
	}
}
