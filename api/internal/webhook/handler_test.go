package webhook

import (
	"bytes"
	"context"
	"encoding/json"
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
