package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/Substrate/api/internal/db"
	"github.com/google/uuid"
)

func TestPushHandler(t *testing.T) {
	mockStore := &db.MockStore{
		UpsertOrgFunc: func(ctx context.Context, installationID int64, orgName string) (uuid.UUID, error) {
			return uuid.New(), nil
		},
		UpsertRepoFunc: func(ctx context.Context, orgID uuid.UUID, githubRepoID int64, name, fullName string) (uuid.UUID, error) {
			return uuid.New(), nil
		},
		UpsertContractFunc: func(ctx context.Context, repoID uuid.UUID, schemaType, specPath, branch, commitSHA, rawContent string) (uuid.UUID, error) {
			return uuid.New(), nil
		},
		UpsertDependencyFunc: func(ctx context.Context, consumerRepoID, providerContractID uuid.UUID, confidenceScore int) error {
			return nil
		},
	}

	payload := PushPayload{
		InstallationID: 123,
		Org:            "myorg",
		Repo:           "myorg/frontend",
		GithubRepoID:   456,
		CommitSHA:      "abcdef",
		Files: []File{
			{
				Path: ".env.example",
				Content: `USERS_API_URL=https://users.myorg.com
PAYMENTS_ENDPOINT="https://api.payments.myorg.com"
DB_URL=postgres://localhost
SECRET_KEY=123`,
			},
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhook", bytes.NewReader(body))
	w := httptest.NewRecorder()

	PushHandler(mockStore).ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var res map[string]int
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Errorf("failed to decode response: %v", err)
	}

	if res["discovered"] != 2 {
		t.Errorf("expected 2 discovered dependencies, got %d", res["discovered"])
	}
}

type mockSyncClient struct {
	syncErr error
}

func (m *mockSyncClient) SyncSchema(ctx context.Context, schema string) error {
	return m.syncErr
}

func TestHandlerWithClient(t *testing.T) {
	tests := []struct {
		name       string
		payload    WebhookPayload
		statusCode int
		invalidJSON bool
		syncErr    error
	}{
		{
			name: "valid safe payload",
			payload: WebhookPayload{
				Event:  "schema_update",
				Status: "SAFE",
				Type:   "openapi",
				Schema: "{}",
			},
			statusCode: http.StatusOK,
		},
		{
			name: "valid approved payload",
			payload: WebhookPayload{
				Event:  "schema_update",
				Status: "APPROVED",
				Type:   "openapi",
				Schema: "{}",
			},
			statusCode: http.StatusOK,
		},
		{
			name: "ignored status",
			payload: WebhookPayload{
				Event:  "schema_update",
				Status: "PENDING",
				Type:   "openapi",
				Schema: "{}",
			},
			statusCode: http.StatusOK,
		},
		{
			name: "ignored type",
			payload: WebhookPayload{
				Event:  "schema_update",
				Status: "SAFE",
				Type:   "graphql",
				Schema: "{}",
			},
			statusCode: http.StatusOK,
		},
		{
			name: "invalid json",
			invalidJSON: true,
			statusCode: http.StatusBadRequest,
		},
		{
			name: "sync error",
			payload: WebhookPayload{
				Event:  "schema_update",
				Status: "SAFE",
				Type:   "openapi",
				Schema: "{}",
			},
			statusCode: http.StatusInternalServerError,
			syncErr:    errors.New("sync failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			if tt.invalidJSON {
				body = []byte("invalid")
			} else {
				body, _ = json.Marshal(tt.payload)
			}
			req := httptest.NewRequest("POST", "/", bytes.NewReader(body))
			w := httptest.NewRecorder()

			client := &mockSyncClient{syncErr: tt.syncErr}
			HandlerWithClient(client).ServeHTTP(w, req)

			if w.Code != tt.statusCode {
				t.Errorf("Expected status %d, got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	handler := Handler()
	if handler == nil {
		t.Error("Expected Handler to return a non-nil http.HandlerFunc")
	}
}
