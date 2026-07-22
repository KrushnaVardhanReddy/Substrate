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

func TestSpecHandler_Success(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/spec/myorg/backend-api", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetPathValue("org", "myorg")
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
	handler := handlers.SpecHandler(mockStore)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to parse json response: %v", err)
	}

	if response["spec"] != "openapi: '3.0.0'" {
		t.Errorf("expected spec 'openapi: '3.0.0'', got %v", response["spec"])
	}
}

func TestSpecHandler_NotFound(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/spec/myorg/unknown-repo", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.SetPathValue("org", "myorg")
	req.SetPathValue("repo", "unknown-repo")

	mockStore := &db.MockStore{
		GetContractsByProviderFullNameFunc: func(ctx context.Context, providerFullName string) ([]db.Contract, error) {
			return []db.Contract{}, nil
		},
	}

	rr := httptest.NewRecorder()
	handler := handlers.SpecHandler(mockStore)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}
}

func TestSpecHandler_MissingParam(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/spec/", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Missing org and repo in SetPathValue
	req.SetPathValue("org", "")
	req.SetPathValue("repo", "")

	mockStore := &db.MockStore{}

	rr := httptest.NewRecorder()
	handler := handlers.SpecHandler(mockStore)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}

	if !strings.Contains(rr.Body.String(), "org and repo are required") {
		t.Errorf("expected error message for missing org and repo, got %v", rr.Body.String())
	}
}
