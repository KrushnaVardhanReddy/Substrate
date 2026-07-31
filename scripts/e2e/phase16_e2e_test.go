package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const p16ApiURL = "http://localhost:8090"

func createP16JWT(org, role string) string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "local-jwt-secret"
	}
	hash := sha256.Sum256([]byte(secret))
	key, _ := paseto.V4SymmetricKeyFromBytes(hash[:])

	token := paseto.NewToken()
	token.Set("orgs", map[string]string{org: role})
	token.SetExpiration(time.Now().Add(time.Hour))

	return token.V4Encrypt(key, nil)
}

func waitForP16Services(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 5; i++ {
		resp, err := client.Get(p16ApiURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	t.Skipf("API server not reachable at %s. Please ensure it is running.", p16ApiURL)
}

func setupP16Database(t *testing.T, proxyBaseURL string) (*pgxpool.Pool, string) {
	ctx := context.Background()
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable"
	}
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err, "Failed to connect to real PostgreSQL")

	_, err = pool.Exec(ctx, "DELETE FROM repo_guides WHERE org = 'e2e-org' AND repo = 'guide-api-repo'")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM repositories WHERE org_id IN (SELECT id FROM organizations WHERE github_org_name = 'e2e-org') AND name = 'guide-api-repo'")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM organizations WHERE github_org_name = 'e2e-org'")
	require.NoError(t, err)

	var orgID string
	err = pool.QueryRow(ctx, "INSERT INTO organizations (github_installation_id, github_org_name) VALUES (1616, 'e2e-org') ON CONFLICT (github_installation_id) DO UPDATE SET github_org_name = EXCLUDED.github_org_name RETURNING id").Scan(&orgID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO repositories (id, org_id, github_repo_id, name, full_name, base_url) VALUES ($1, $2, 161616, 'guide-api-repo', 'e2e-org/guide-api-repo', $3)", uuid.New().String(), orgID, proxyBaseURL)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO repo_guides (org, repo, file_path, title, content) VALUES ('e2e-org', 'guide-api-repo', 'docs/authentication.md', 'Authentication Guide', '# Auth\nUse Bearer tokens.') ON CONFLICT DO NOTHING")
	require.NoError(t, err)

	return pool, orgID
}

func TestPhase16_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	waitForP16Services(t)

	mockProxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/test-endpoint" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"proxy_success": true}`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer mockProxy.Close()

	pool, _ := setupP16Database(t, mockProxy.URL)
	defer pool.Close()

	adminJWT := createP16JWT("e2e-org", "admin")

	var sandboxToken string
	tests := []struct {
		name           string
		method         string
		url            string
		body           interface{}
		auth           func() string
		expectedStatus int
		assertResponse func(t *testing.T, respBody []byte)
	}{
		{
			name:           "GET /api/v1/docs/e2e-org/guide-api-repo",
			method:         "GET",
			url:            p16ApiURL + "/api/v1/docs/e2e-org/guide-api-repo",
			body:           nil,
			auth:           func() string { return "" },
			expectedStatus: http.StatusOK,
			assertResponse: func(t *testing.T, respBody []byte) {
				var guides []map[string]interface{}
				require.NoError(t, json.Unmarshal(respBody, &guides))
				assert.Len(t, guides, 1)
				assert.Equal(t, "Authentication Guide", guides[0]["title"])
			},
		},
		{
			name:           "GET /api/v1/docs/e2e-org/guide-api-repo/docs/authentication.md",
			method:         "GET",
			url:            p16ApiURL + "/api/v1/docs/e2e-org/guide-api-repo/docs/authentication.md",
			body:           nil,
			auth:           func() string { return "" },
			expectedStatus: http.StatusOK,
			assertResponse: func(t *testing.T, respBody []byte) {
				var guide map[string]interface{}
				require.NoError(t, json.Unmarshal(respBody, &guide))
				assert.Equal(t, "Authentication Guide", guide["title"])
				assert.Contains(t, guide["content"].(string), "Use Bearer tokens")
			},
		},
		{
			name:           "POST /api/v1/sandbox/token",
			method:         "POST",
			url:            p16ApiURL + "/api/v1/org/e2e-org/sandbox/token",
			body:           map[string]string{"org": "e2e-org", "repo": "guide-api-repo"},
			auth:           func() string { return adminJWT },
			expectedStatus: http.StatusOK,
			assertResponse: func(t *testing.T, respBody []byte) {
				var res map[string]string
				require.NoError(t, json.Unmarshal(respBody, &res))
				sandboxToken = res["token"]
				assert.NotEmpty(t, sandboxToken)
			},
		},
		{
			name:           "POST /api/v1/sandbox/request (Authenticated)",
			method:         "POST",
			url:            p16ApiURL + "/api/v1/sandbox/request",
			body: map[string]interface{}{
				"org":     "e2e-org",
				"repo":    "guide-api-repo",
				"method":  "GET",
				"path":    "/test-endpoint",
				"headers": map[string]string{},
				"body":    "",
			},
			auth:           func() string { return sandboxToken },
			expectedStatus: http.StatusOK,
			assertResponse: func(t *testing.T, respBody []byte) {
				var resBody map[string]interface{}
				require.NoError(t, json.Unmarshal(respBody, &resBody))
				assert.Equal(t, float64(200), resBody["status_code"])
				assert.Contains(t, resBody["body"].(string), "proxy_success")
			},
		},
		{
			name:           "POST /api/v1/sandbox/request (Unauthenticated)",
			method:         "POST",
			url:            p16ApiURL + "/api/v1/sandbox/request",
			body: map[string]interface{}{
				"org":     "e2e-org",
				"repo":    "guide-api-repo",
				"method":  "GET",
				"path":    "/test-endpoint",
				"headers": map[string]string{},
				"body":    "",
			},
			auth:           func() string { return "" },
			expectedStatus: http.StatusUnauthorized,
			assertResponse: func(t *testing.T, respBody []byte) {},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var reqBodyBytes []byte
			if tc.body != nil {
				reqBodyBytes, _ = json.Marshal(tc.body)
			}
			req, err := http.NewRequest(tc.method, tc.url, bytes.NewReader(reqBodyBytes))
			require.NoError(t, err)

			if tc.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			auth := tc.auth()
			if auth != "" {
				req.Header.Set("Authorization", "Bearer "+auth)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			if tc.assertResponse != nil {
				respBody, err := io.ReadAll(resp.Body)
				require.NoError(t, err)
				if resp.StatusCode != tc.expectedStatus {
					t.Logf("Debug Response Body: %q", string(respBody))
				}
				tc.assertResponse(t, respBody)
			}
		})
	}
}
