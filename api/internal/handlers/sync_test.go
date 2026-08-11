// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/services"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/workers"
)

func TestSyncHandler_MissingAuthToken(t *testing.T) {
	req, err := http.NewRequest("POST", "/api/v1/sync", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	mockEnqueuer := &workers.MockJobEnqueuer{}
	handler := SyncHandler(&db.MockStore{}, mockEnqueuer)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusUnauthorized)
	}
}

func TestSyncHandler_InvalidJSON(t *testing.T) {
	os.Setenv("INTERNAL_SERVICE_TOKEN", "secret")
	defer os.Unsetenv("INTERNAL_SERVICE_TOKEN")

	req, err := http.NewRequest("POST", "/api/v1/sync", bytes.NewBufferString(`{invalid json`))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer secret")

	rr := httptest.NewRecorder()
	mockEnqueuer := &workers.MockJobEnqueuer{}
	handler := SyncHandler(&db.MockStore{}, mockEnqueuer)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestSyncHandler_ValidRequest(t *testing.T) {
	os.Setenv("INTERNAL_SERVICE_TOKEN", "secret")
	defer os.Unsetenv("INTERNAL_SERVICE_TOKEN")

	payload := services.SyncRequest{
		InstallationID:       123456,
		Org:                  "myorg",
		ConsumerRepo:         "myorg/frontend",
		ConsumerGithubRepoID: 789,
		CommitSHA:            "abc1234",
		Dependencies: []services.DependencyPayload{
			{
				ProviderRepo:         "myorg/backend-api",
				ProviderGithubRepoID: 456,
				SchemaType:           "openapi",
				SpecPath:             "api/openapi.yaml",
				Branch:               "main",
				RawContent:           "openapi: 3.0.0",
			},
		},
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "/api/v1/sync", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer secret")

	mockStore := &db.MockStore{}
	rr := httptest.NewRecorder()
	mockEnqueuer := &workers.MockJobEnqueuer{}
	handler := SyncHandler(mockStore, mockEnqueuer)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusAccepted)
	}
}
