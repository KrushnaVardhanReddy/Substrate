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

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/services"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/workers"
	"github.com/google/uuid"
)

func TestCrossRepoCheckHandler_MissingAuthToken(t *testing.T) {
	req, err := http.NewRequest("POST", "/api/v1/cross-repo-check", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	mockEnqueuer := &workers.MockJobEnqueuer{}
	handler := CrossRepoCheckHandler(&db.MockStore{}, mockEnqueuer)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestCrossRepoCheckHandler_ValidRequest_0Consumers(t *testing.T) {
	os.Setenv("INTERNAL_SERVICE_TOKEN", "secret")
	defer os.Unsetenv("INTERNAL_SERVICE_TOKEN")

	payload := services.CrossRepoCheckRequest{
		InstallationID:    123456,
		Org:               "myorg",
		ProviderRepo:      "myorg/backend-api",
		HeadSchemaContent: "openapi: 3.0.0",
		SchemaType:        "openapi",
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "/api/v1/cross-repo-check", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer secret")

	mockStore := &db.MockStore{
		GetContractsByProviderFullNameFunc: func(ctx context.Context, providerFullName string) ([]db.Contract, error) {
			return []db.Contract{{ID: uuid.New(), SchemaType: "openapi"}}, nil
		},
		GetConsumersByProviderContractFunc: func(ctx context.Context, providerContractID uuid.UUID) ([]db.ConsumerDependency, error) {
			return []db.ConsumerDependency{}, nil
		},
	}

	rr := httptest.NewRecorder()
	mockEnqueuer := &workers.MockJobEnqueuer{}
	handler := CrossRepoCheckHandler(mockStore, mockEnqueuer)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusAccepted {
		t.Fatalf("expected status 202, got %d", status)
	}
}

func TestCrossRepoCheckHandler_ValidRequest_1BrokenConsumer(t *testing.T) {
	os.Setenv("INTERNAL_SERVICE_TOKEN", "secret")
	defer os.Unsetenv("INTERNAL_SERVICE_TOKEN")

	// Mock diff engine
	diffEngine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := services.DiffReport{
			Summary: services.DiffReportSummary{BreakingCount: 1},
			Breaking: []interface{}{
				map[string]interface{}{"rule": "ENDPOINT_MODIFIED"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer diffEngine.Close()

	os.Setenv("DIFF_ENGINE_URL", diffEngine.URL)
	defer os.Unsetenv("DIFF_ENGINE_URL")

	payload := services.CrossRepoCheckRequest{
		InstallationID:    123456,
		Org:               "myorg",
		ProviderRepo:      "myorg/backend-api",
		HeadSchemaContent: "openapi: 3.0.0",
		SchemaType:        "openapi",
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", "/api/v1/cross-repo-check", bytes.NewBuffer(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer secret")

	mockStore := &db.MockStore{
		GetContractsByProviderFullNameFunc: func(ctx context.Context, providerFullName string) ([]db.Contract, error) {
			return []db.Contract{{ID: uuid.New(), SchemaType: "openapi"}}, nil
		},
		GetConsumersByProviderContractFunc: func(ctx context.Context, providerContractID uuid.UUID) ([]db.ConsumerDependency, error) {
			return []db.ConsumerDependency{
				{ConsumerFullName: "myorg/frontend", ContractRawContent: "openapi: 3.0.0 base"},
			}, nil
		},
	}

	rr := httptest.NewRecorder()
	mockEnqueuer := &workers.MockJobEnqueuer{}
	handler := CrossRepoCheckHandler(mockStore, mockEnqueuer)

	handler.ServeHTTP(rr, req)
}
