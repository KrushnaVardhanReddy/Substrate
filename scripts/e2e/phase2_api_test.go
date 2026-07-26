package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	p2ApiURL    = "http://localhost:8090"
	p2DbURL     = "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
	p2JWTSecret = "local-jwt-secret"
)

func waitForServicesP2(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	var apiReady, dbReady bool

	for i := 0; i < 20; i++ {
		if !apiReady {
			resp, err := client.Get(p2ApiURL + "/health")
			if err == nil && resp.StatusCode == http.StatusOK {
				apiReady = true
				resp.Body.Close()
			}
		}

		if !dbReady {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			pool, err := pgxpool.New(ctx, p2DbURL)
			if err == nil {
				if err := pool.Ping(ctx); err == nil {
					dbReady = true
				}
				pool.Close()
			}
			cancel()
		}

		if apiReady && dbReady {
			return
		}
		time.Sleep(1 * time.Second)
	}

	if !apiReady || !dbReady {
		t.Skipf("Live infrastructure unreachable (API: %v, DB: %v). Skipping E2E test.", apiReady, dbReady)
	}
}

func TestPhase2GitHubAppE2E(t *testing.T) {
	waitForServicesP2(t)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, p2DbURL)
	require.NoError(t, err)
	defer pool.Close()

	// Seed organizations and repositories if not present
	_, err = pool.Exec(ctx, `
		INSERT INTO organizations (id, github_installation_id, github_org_name) VALUES
		('11111111-1111-1111-1111-111111111111', 123456, 'mcp-org')
		ON CONFLICT (id) DO NOTHING;
	`)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		INSERT INTO repositories (id, org_id, github_repo_id, name, full_name) VALUES
		('22222222-2222-2222-2222-222222222222', '11111111-1111-1111-1111-111111111111', 101, 'backend', 'mcp-org/backend'),
		('33333333-3333-3333-3333-333333333333', '11111111-1111-1111-1111-111111111111', 102, 'frontend', 'mcp-org/frontend'),
		('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 103, 'discovery-test-repo', 'mcp-org/discovery-test-repo')
		ON CONFLICT (id) DO NOTHING;
	`)
	require.NoError(t, err)

	// Create a mock GitHub server for outbound requests from the webhook handler
	mockGitHub := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer mockGitHub.Close()

	t.Run("POST /api/v1/webhook handles GitHub Push Event with HMAC signature", func(t *testing.T) {
		// Construct a mock GitHub push event JSON payload
		payload := map[string]interface{}{
			"ref": "refs/heads/main",
			"repository": map[string]interface{}{
				"id":        101,
				"name":      "backend",
				"full_name": "mcp-org/backend",
				"owner": map[string]interface{}{
					"login": "mcp-org",
				},
			},
			"installation": map[string]interface{}{
				"id": 123456,
			},
			"commits": []map[string]interface{}{
				{
					"id": "mock-commit-sha",
					"added": []string{
						"openapi.yaml",
					},
					"modified": []string{},
					"removed":  []string{},
				},
			},
		}

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, p2ApiURL+"/api/v1/webhook", bytes.NewBuffer(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-GitHub-Event", "push")

		// In GitHub Apps, the webhook payload is verified using an HMAC signature with the webhook secret.
		// Assuming 'local-jwt-secret' or empty is used for tests, we generate the signature.
		// NOTE: Often webhook secrets are specifically loaded from the env, but in sandbox `local-jwt-secret` is the default service token.
		// Wait, looking at the code, it uses serviceTokenMW for the webhook, which just checks the service token. Let's provide it in the Authorization header as well to be safe if that's what the MW wants.
		// Let's set both a mock X-Hub-Signature-256 and the service token just in case.
		mac := hmac.New(sha256.New, []byte(p2JWTSecret))
		mac.Write(body)
		signature := hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Hub-Signature-256", "sha256="+signature)

		req.Header.Set("Authorization", "Bearer local-dev-token")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		// Based on standard behaviors for a webhook endpoint, expect 200 OK or 202 Accepted
		assert.Contains(t, []int{http.StatusOK, http.StatusAccepted, http.StatusNoContent}, resp.StatusCode, "webhook endpoint should accept the payload")
	})
}
