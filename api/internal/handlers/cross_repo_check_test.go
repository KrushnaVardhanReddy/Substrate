package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/KrushnaVardhanReddy/Substrate/api/internal/db"
	"github.com/google/uuid"
)

func TestCrossRepoCheckHandler_MissingAuthToken(t *testing.T) {
	req, err := http.NewRequest("POST", "/api/v1/cross-repo-check", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := CrossRepoCheckHandler(&db.MockStore{})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusUnauthorized {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
	}
}

func TestCrossRepoCheckHandler_ValidRequest_0Consumers(t *testing.T) {
	os.Setenv("INTERNAL_SERVICE_TOKEN", "secret")
	defer os.Unsetenv("INTERNAL_SERVICE_TOKEN")

	payload := CrossRepoCheckRequest{
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
	handler := CrossRepoCheckHandler(mockStore)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}

	var resp CrossRepoCheckResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if resp.TotalConsumers != 0 {
		t.Errorf("expected 0 total consumers, got %d", resp.TotalConsumers)
	}
	if !resp.IsSafe {
		t.Errorf("expected is_safe to be true, got %v", resp.IsSafe)
	}
}

func TestCrossRepoCheckHandler_ValidRequest_1BrokenConsumer(t *testing.T) {
	os.Setenv("INTERNAL_SERVICE_TOKEN", "secret")
	defer os.Unsetenv("INTERNAL_SERVICE_TOKEN")

	// Mock diff engine
	diffEngine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := DiffReport{
			Summary: DiffReportSummary{BreakingCount: 1},
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

	payload := CrossRepoCheckRequest{
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
	handler := CrossRepoCheckHandler(mockStore)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}

	var resp CrossRepoCheckResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if resp.TotalConsumers != 1 {
		t.Errorf("expected 1 total consumer, got %d", resp.TotalConsumers)
	}
	if resp.BrokenConsumers != 1 {
		t.Errorf("expected 1 broken consumer, got %d", resp.BrokenConsumers)
	}
	if resp.IsSafe {
		t.Errorf("expected is_safe to be false, got %v", resp.IsSafe)
	}
	if len(resp.Results) != 1 || resp.Results[0].Status != "breaking" {
		t.Errorf("expected 1 breaking result, got %v", resp.Results)
	}
}
