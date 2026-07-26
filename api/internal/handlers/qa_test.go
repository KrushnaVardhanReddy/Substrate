package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

func TestQAPostmanHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/v1/qa/postman/mcp-org/shadow-api-repo", nil)
	rr := httptest.NewRecorder()

	mockStore := &db.MockStore{
		GetOrgIDByNameFunc: func(ctx context.Context, orgName string) (uuid.UUID, error) {
			return uuid.New(), nil
		},
	}

	r := chi.NewRouter()
	r.Get("/api/v1/qa/postman/{org}/{repo}", QAPostmanHandler(mockStore))
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)

	info, ok := resp["info"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "Auto-Generated Collection for mcp-org/shadow-api-repo", info["name"])
}

func TestQAShadowReplayHandler(t *testing.T) {
	reqBody := `{"org":"mcp-org", "repo":"shadow-api-repo", "timestamp":"2026-07-01"}`
	req, _ := http.NewRequest("POST", "/api/v1/qa/shadow/replay", bytes.NewBufferString(reqBody))
	rr := httptest.NewRecorder()

	mockStore := &db.MockStore{
		GetOrgIDByNameFunc: func(ctx context.Context, orgName string) (uuid.UUID, error) {
			return uuid.New(), nil
		},
	}

	r := chi.NewRouter()
	r.Post("/api/v1/qa/shadow/replay", QAShadowReplayHandler(mockStore))
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

func TestQACoverageHandler(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/v1/qa/coverage/mcp-org/shadow-api-repo", nil)
	rr := httptest.NewRecorder()

	mockStore := &db.MockStore{
		GetOrgIDByNameFunc: func(ctx context.Context, orgName string) (uuid.UUID, error) {
			return uuid.New(), nil
		},
	}

	r := chi.NewRouter()
	r.Get("/api/v1/qa/coverage/{org}/{repo}", QACoverageHandler(mockStore))
	r.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)

	assert.Equal(t, 85.5, resp["score"])
}
