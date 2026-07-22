package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	p8RegistryAPIToken = "local-dev-token"
	p8JWTSecret        = "local-jwt-secret"
	p8ApiURL           = "http://localhost:8090"
	p8DbURL            = "postgres://postgres:postgres@localhost:5432/substrate?sslmode=disable"
)

func waitForP8Services(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 5; i++ {
		resp, err := client.Get(p8ApiURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("API server not reachable at %s. Please ensure 'make api' and 'make postgres' are running.", p8ApiURL)
}

func setupP8Database(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, p8DbURL)
	require.NoError(t, err, "Failed to connect to real PostgreSQL")

	// Clean up tables relevant to Phase 8
	_, err = pool.Exec(ctx, "DELETE FROM river_job")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM org_webhooks")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM dependencies")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM contracts")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM repositories")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM organizations")
	require.NoError(t, err)

	return pool
}

func createJWT(org, role string) string {
	claims := jwt.MapClaims{
		"orgs": map[string]interface{}{
			org: role,
		},
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(p8JWTSecret))
	return tokenString
}

func TestPhase8SystemE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	ctx := context.Background()

	waitForP8Services(t)

	// Build the CLI binary
	engineDir, err := filepath.Abs("../../engine")
	require.NoError(t, err)
	binPath := filepath.Join(engineDir, "substrate_p8_test_bin")
	cmdBuild := exec.Command("go", "build", "-o", "substrate_p8_test_bin", "./cmd/substrate")
	cmdBuild.Dir = engineDir
	require.NoError(t, cmdBuild.Run(), "Failed to compile the substrate CLI")
	defer os.Remove(binPath) // Cleanup

	// Database operations for setup
	pool := setupP8Database(t)
	defer pool.Close()

	var orgID string
	err = pool.QueryRow(ctx, "INSERT INTO organizations (github_installation_id, github_org_name) VALUES (999, 'acme') RETURNING id").Scan(&orgID)
	require.NoError(t, err)

	var billingApiRepoID string
	err = pool.QueryRow(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name) VALUES ($1, 111, 'billing-api', 'acme/billing-api') RETURNING id", orgID).Scan(&billingApiRepoID)
	require.NoError(t, err)

	var invoiceSvcRepoID string
	err = pool.QueryRow(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name) VALUES ($1, 222, 'invoice-service', 'acme/invoice-service') RETURNING id", orgID).Scan(&invoiceSvcRepoID)
	require.NoError(t, err)

	t.Run("Scenario 1: Durable Job Queue & Webhook Egress (P8-T01)", func(t *testing.T) {
		webhookReceived := make(chan bool, 1)
		var hmacSig string

		// Mock target server
		targetServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hmacSig = r.Header.Get("X-Hub-Signature-256")
			w.WriteHeader(http.StatusOK)
			webhookReceived <- true
		}))
		defer targetServer.Close()

		// Register webhook directly into database since we can't import internal/db easily
		_, err = pool.Exec(ctx, "INSERT INTO org_webhooks (org, url, secret) VALUES ($1, $2, $3)", "acme", targetServer.URL, "my-webhook-secret")
		require.NoError(t, err)

		// Submit a breaking change diff to trigger the event
		diffReq := map[string]interface{}{
			"diff_report":         json.RawMessage(`{"summary":{"breaking_count":1},"breaking_changes":[{"description":"Breaking change detected"}]}`),
			"is_audit_mode":       false,
			"org":                 "acme",
			"provider_repo":       "acme/billing-api",
			"pr_number":           42,
			"commit_sha":          "abc123sha",
			"head_schema_content": "openapi: 3.0.0\ninfo:\n  version: 2.0.0",
			"schema_type":         "openapi",
		}
		body, _ := json.Marshal(diffReq)

		req, err := http.NewRequest("POST", p8ApiURL+"/api/v1/diff", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+p8RegistryAPIToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// Wait for job queue processing (async background worker)
		select {
		case <-webhookReceived:
			require.NotEmpty(t, hmacSig)
			assert.True(t, strings.HasPrefix(hmacSig, "sha256="), "HMAC signature should start with sha256=")
		case <-time.After(10 * time.Second):
			// Check river_job table to see if it failed
			var state, errs string
			pool.QueryRow(ctx, "SELECT state, errors FROM river_job WHERE kind = 'egress_webhook'").Scan(&state, &errs)
			t.Fatalf("Timeout waiting for Webhook Egress worker. River job state: %s, errors: %s", state, errs)
		}
	})

	t.Run("Scenario 2: Enterprise Authz & RBAC (P8-T02)", func(t *testing.T) {
		viewerJWT := createJWT("acme", "Read-Only")
		adminJWT := createJWT("acme", "admin") // Note: The AuthzMiddleware expects exactly "admin"

		// Use the Viewer JWT
		reqViewer, _ := http.NewRequest("POST", p8ApiURL+"/api/v1/org/acme/webhooks", bytes.NewReader([]byte(`{"url":"http://test","secret":"test"}`)))
		reqViewer.Header.Set("Authorization", "Bearer "+viewerJWT)
		reqViewer.Header.Set("Content-Type", "application/json")

		respViewer, err := http.DefaultClient.Do(reqViewer)
		require.NoError(t, err)
		defer respViewer.Body.Close()
		assert.Equal(t, http.StatusForbidden, respViewer.StatusCode, "Read-Only role should get 403 Forbidden")

		// Use the Admin JWT
		reqAdmin, _ := http.NewRequest("POST", p8ApiURL+"/api/v1/org/acme/webhooks", bytes.NewReader([]byte(`{"url":"http://test","secret":"test"}`)))
		reqAdmin.Header.Set("Authorization", "Bearer "+adminJWT)
		reqAdmin.Header.Set("Content-Type", "application/json")

		respAdmin, err := http.DefaultClient.Do(reqAdmin)
		require.NoError(t, err)
		defer respAdmin.Body.Close()
		assert.Equal(t, http.StatusCreated, respAdmin.StatusCode, "Admin role should get 201 Created")
	})

	t.Run("Scenario 3: Cascading Rollback Gate (P8-T03)", func(t *testing.T) {
		// Insert billing-api at v2
		contractV2Content := `openapi: 3.0.0
info:
  title: Billing API
  version: 2.0.0
paths:
  /invoices:
    post:
      responses:
        '200':
          description: OK
`
		var v2ContractID string
		err = pool.QueryRow(ctx, "INSERT INTO contracts (repo_id, schema_type, spec_path, branch, latest_commit_sha, raw_content) VALUES ($1, 'openapi', 'openapi.yaml', 'main', 'v2-sha', $2) RETURNING id", billingApiRepoID, contractV2Content).Scan(&v2ContractID)
		require.NoError(t, err)

		// Consumer depends on v2
		_, err = pool.Exec(ctx, "INSERT INTO dependencies (consumer_repo_id, provider_contract_id, confidence_score) VALUES ($1, $2, 100)", invoiceSvcRepoID, v2ContractID)
		require.NoError(t, err)

		// Insert v1-sha to exist in the database so it can be evaluated
		contractV1Content := `openapi: 3.0.0
info:
  title: Billing API
  version: 1.0.0
paths: {}
`
		_, err = pool.Exec(ctx, "INSERT INTO contracts (repo_id, schema_type, spec_path, branch, latest_commit_sha, raw_content) VALUES ($1, 'openapi', 'openapi.yaml', 'v1-branch', 'v1-sha', $2)", billingApiRepoID, contractV1Content)
		require.NoError(t, err)

		cmd := exec.Command(binPath, "check-rollback", "--repo", "acme/billing-api", "--target-sha", "v1-sha")
		cmd.Env = append(os.Environ(), "SUBSTRATE_API_URL="+p8ApiURL, "REGISTRY_API_TOKEN="+p8RegistryAPIToken)

		var outBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &outBuf

		err = cmd.Run()
		require.Error(t, err)
		if exitError, ok := err.(*exec.ExitError); ok {
			assert.Equal(t, 2, exitError.ExitCode(), "Exit code should be 2 for blocked rollback")
		} else {
			t.Fatalf("Expected ExitError, got %v", err)
		}

		assert.Contains(t, outBuf.String(), "ROLLBACK BLOCKED")
		assert.Contains(t, outBuf.String(), "acme/invoice-service")
	})

	t.Run("Scenario 4: Billing Paywall Pause (Graceful Degradation) (P8-T07)", func(t *testing.T) {
		// Update trial_ends_at to the past
		_, err := pool.Exec(ctx, "UPDATE organizations SET trial_ends_at = NOW() - INTERVAL '10 days', stripe_customer_id = NULL WHERE github_org_name = 'acme'")
		require.NoError(t, err)

		// Create a mock push handler payload
		pushPayload := map[string]interface{}{
			"installation_id": 999,
			"org":             "acme",
			"repo":            "billing-api",
			"commit_sha":      "commitsha123",
			"branch":          "main",
		}
		body, _ := json.Marshal(pushPayload)

		req, err := http.NewRequest("POST", p8ApiURL+"/api/v1/webhook", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+p8RegistryAPIToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusAccepted, resp.StatusCode, "Push webhook should be gracefully paused with 202 Accepted")

		var respBody map[string]string
		json.NewDecoder(resp.Body).Decode(&respBody)
		assert.Equal(t, "paused_due_to_billing", respBody["status"])
	})
}
