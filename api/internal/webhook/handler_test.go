package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/services"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/workers"
	"github.com/google/uuid"
	"os"
)

func TestPushHandler(t *testing.T) {
	contractID := uuid.New()
	consumerRepoID := uuid.New()

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
		GetContractsByProviderFullNameFunc: func(ctx context.Context, providerFullName string) ([]db.Contract, error) {
			return []db.Contract{
				{
					ID:         contractID,
					SpecPath:   "openapi.yaml",
					RawContent: "old schema",
				},
			}, nil
		},
		GetConsumersByProviderContractFunc: func(ctx context.Context, providerContractID uuid.UUID) ([]db.ConsumerDependency, error) {
			return []db.ConsumerDependency{
				{
					ConsumerRepoID:   consumerRepoID,
					ConsumerFullName: "consumer-org/consumer-repo",
				},
			}, nil
		},
	}

	payload := services.PushPayload{
		InstallationID: 123,
		Org:            "myorg",
		Repo:           "myorg/frontend",
		GithubRepoID:   456,
		CommitSHA:      "abcdef",
		Files: []services.File{
			{
				Path:    "openapi.yaml",
				Content: "new schema",
			},
		},
	}

	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/webhook", bytes.NewReader(body))
	w := httptest.NewRecorder()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"summary": {"breaking_count": 1},
			"breaking_changes": [{"severity": "BREAKING", "description": "test breaking change", "rule_id": "test-rule", "path": "test-path"}]
		}`))
	}))
	defer ts.Close()
	os.Setenv("DIFF_ENGINE_URL", ts.URL)
	defer os.Unsetenv("DIFF_ENGINE_URL")

	prCreated := make(chan bool, 1)

	mockGHClient := &github.MockClient{
		CreateDraftPRFunc: func(ctx context.Context, owner, repo, branch, patch, title, body string) (string, error) {
			if owner != "consumer-org" || repo != "consumer-repo" {
				t.Errorf("expected PR for consumer-org/consumer-repo, got %s/%s", owner, repo)
			}
			if title != "chore(substrate): Auto-fix breaking change from upstream [myorg/frontend]" {
				t.Errorf("unexpected PR title: %s", title)
			}
			prCreated <- true
			return "https://github.com/mock/pull/1", nil
		},
		SearchCodeFunc: func(ctx context.Context, owner, repo, query string) (string, error) {
			return "src/index.ts", nil
		},
		GetFileContentFunc: func(ctx context.Context, owner, repo, path string) (string, error) {
			return "const x = 1;", nil
		},
	}

	// For tests we would typically mock the river client, but since we refactored it
	// we will inject a dummy client or just test the returned response
	mockEnqueuer := &workers.MockJobEnqueuer{}
	PushHandler(mockStore, mockGHClient, mockEnqueuer).ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Errorf("expected status %d, got %d", http.StatusAccepted, w.Code)
	}

	// We no longer expect CreateDraftPR to be called synchronously since it's enqueued in River.
	// The PR creation logic moved to `services.ProcessPush` which is called by the background worker.
	// Therefore we don't block on `prCreated` channel anymore in the handler test.
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
