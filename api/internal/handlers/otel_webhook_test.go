package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestOTelMetricsHandler(t *testing.T) {
	mockStore := &db.MockStore{}
	handler := OTelMetricsHandler(mockStore)

	repoID := uuid.New()
	now := time.Now()

	spans := []OTelMetricsSpan{
		{
			RepoID:    repoID,
			Method:    "GET",
			Path:      "/api/v1/users",
			Timestamp: now,
		},
	}

	body, err := json.Marshal(spans)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/telemetry/traffic", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusAccepted, rr.Code)

	var resp map[string]string
	err = json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "accepted", resp["status"])
}

func TestGetZombiesHandler(t *testing.T) {
	mockStore := &db.MockStore{}
	handler := GetZombiesHandler(mockStore)

	req, err := http.NewRequest(http.MethodGet, "/api/v1/org/test-org/zombies", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Get("/api/v1/org/{org}/zombies", handler)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestCreateZombiePRHandler(t *testing.T) {
	mockClient := &github.MockClient{}
	handler := CreateZombiePRHandler(mockClient)

	reqBody := ZombiePRRequest{
		RepoName: "test-repo",
		Method:   "GET",
		Path:     "/api/v1/old",
	}
	body, err := json.Marshal(reqBody)
	assert.NoError(t, err)

	req, err := http.NewRequest(http.MethodPost, "/api/v1/org/test-org/zombies/pr", bytes.NewBuffer(body))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	r := chi.NewRouter()
	r.Post("/api/v1/org/{org}/zombies/pr", handler)
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]string
	err = json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "https://github.com/mock/mock/pull/1", resp["pr_url"])
}
