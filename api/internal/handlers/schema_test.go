// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/handlers"
)

func TestSchemaHandler_Success(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/schema/myorg/backend-api", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetPathValue("owner", "myorg")
	req.SetPathValue("repo", "backend-api")

	mockStore := &db.MockStore{
		GetContractsByProviderFullNameFunc: func(ctx context.Context, providerFullName string) ([]db.Contract, error) {
			if providerFullName == "myorg/backend-api" {
				return []db.Contract{
					{
						SchemaType:      "openapi",
						SpecPath:        "openapi.yaml",
						RawContent:      "openapi: '3.0.0'",
						LatestCommitSHA: "abc123",
						SyncedAt:        time.Now(),
					},
				}, nil
			}
			return nil, nil
		},
	}

	rr := httptest.NewRecorder()
	handler := handlers.SchemaHandler(mockStore)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if response["schema"] != "openapi: '3.0.0'" {
		t.Errorf("expected schema 'openapi: '3.0.0'', got %v", response["schema"])
	}

	if response["spec_path"] != "openapi.yaml" {
		t.Errorf("expected spec_path 'openapi.yaml', got %v", response["spec_path"])
	}
}

func TestSchemaHandler_NotFound(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/schema/myorg/unknown-repo", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetPathValue("owner", "myorg")
	req.SetPathValue("repo", "unknown-repo")

	mockStore := &db.MockStore{
		GetContractsByProviderFullNameFunc: func(ctx context.Context, providerFullName string) ([]db.Contract, error) {
			return []db.Contract{}, nil
		},
	}

	rr := httptest.NewRecorder()
	handler := handlers.SchemaHandler(mockStore)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestSchemaHandler_MissingParam(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/schema/", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Missing owner and repo in SetPathValue
	req.SetPathValue("owner", "")
	req.SetPathValue("repo", "")

	mockStore := &db.MockStore{}

	rr := httptest.NewRecorder()
	handler := handlers.SchemaHandler(mockStore)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	if !strings.Contains(rr.Body.String(), "owner and repo are required") {
		t.Errorf("expected error message for missing owner and repo, got %v", rr.Body.String())
	}
}
