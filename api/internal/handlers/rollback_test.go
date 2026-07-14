package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/google/uuid"
)

func TestCanRollbackHandler_MissingParams(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/registry/can-rollback", nil)
	w := httptest.NewRecorder()

	store := &db.MockStore{}
	handler := CanRollbackHandler(store)
	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request, got %d", w.Result().StatusCode)
	}
}

func TestCanRollbackHandler_TargetNotFound(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/v1/registry/can-rollback?org=org&repo=repo&target_sha=abc", nil)
	w := httptest.NewRecorder()

	store := &db.MockStore{
		GetContractsByProviderFullNameFunc: func(ctx context.Context, providerFullName string) ([]db.Contract, error) {
			return []db.Contract{}, nil
		},
	}

	handler := CanRollbackHandler(store)
	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Result().StatusCode)
	}

	var resp CanRollbackResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if !resp.CanRollback {
		t.Errorf("Expected CanRollback to be true")
	}
}

func TestCanRollbackHandler_SafeRollback(t *testing.T) {
	os.Setenv("DIFF_ENGINE_URL", "http://localhost:8080") // Fallback in perform cross repo check

	req := httptest.NewRequest("GET", "/api/v1/registry/can-rollback?org=org&repo=repo&target_sha=abc", nil)
	w := httptest.NewRecorder()

	store := &db.MockStore{
		GetContractsByProviderFullNameFunc: func(ctx context.Context, providerFullName string) ([]db.Contract, error) {
			return []db.Contract{
				{ID: uuid.New(), LatestCommitSHA: "abc", RawContent: "openapi: 3.0.0\ninfo:\n  version: 1.0.0", SchemaType: "openapi"},
			}, nil
		},
		GetConsumersByProviderContractFunc: func(ctx context.Context, providerContractID uuid.UUID) ([]db.ConsumerDependency, error) {
			return []db.ConsumerDependency{}, nil
		},
	}

	diffEngine := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := DiffReport{
			Summary: DiffReportSummary{BreakingCount: 0},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer diffEngine.Close()
	os.Setenv("DIFF_ENGINE_URL", diffEngine.URL)
	defer os.Unsetenv("DIFF_ENGINE_URL")

	handler := CanRollbackHandler(store)
	handler.ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", w.Result().StatusCode)
	}

	var resp CanRollbackResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if !resp.CanRollback {
		t.Errorf("Expected CanRollback to be true")
	}
}
