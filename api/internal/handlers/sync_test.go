package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

func TestSyncHandler_MissingAuthToken(t *testing.T) {
	req, err := http.NewRequest("POST", "/api/v1/sync", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := SyncHandler(&db.MockStore{})

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
	handler := SyncHandler(&db.MockStore{})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}

func TestSyncHandler_ValidRequest(t *testing.T) {
	os.Setenv("INTERNAL_SERVICE_TOKEN", "secret")
	defer os.Unsetenv("INTERNAL_SERVICE_TOKEN")

	payload := SyncRequest{
		InstallationID:       123456,
		Org:                  "myorg",
		ConsumerRepo:         "myorg/frontend",
		ConsumerGithubRepoID: 789,
		CommitSHA:            "abc1234",
		Dependencies: []DependencyPayload{
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

	rr := httptest.NewRecorder()
	handler := SyncHandler(&db.MockStore{})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]int
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatal(err)
	}

	if response["synced"] != 1 {
		t.Errorf("expected synced count 1, got %v", response["synced"])
	}
}
