package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
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
	hash := sha256.Sum256([]byte(p8JWTSecret))
	key, _ := paseto.V4SymmetricKeyFromBytes(hash[:])

	token := paseto.NewToken()
	token.Set("orgs", map[string]string{org: role})
	token.SetExpiration(time.Now().Add(time.Hour))

	return token.V4Encrypt(key, nil)
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
      summary: Create an invoice
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Invoice'
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Invoice'
components:
  schemas:
    Invoice:
      type: object
      required:
        - amount
      properties:
        id:
          type: string
        amount:
          type: number
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
paths:
  /invoices:
    post:
      summary: Create an invoice
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              # Missing 'amount' property intentionally, making rollback to v1 a breaking change 
              # for consumers who expect 'amount' in V2. Wait, if V1 misses it, and V2 adds it,
              # V1 is breaking because consumer expects the schema to have it.
              properties:
                id:
                  type: string
      responses:
        '200':
          description: OK
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

	t.Run("Scenario 5: RBAC Breadth", func(t *testing.T) {
		viewerJWT := createJWT("rbac-org", "Read-Only")
		adminJWT := createJWT("rbac-org", "admin")

		// Scenario 1: Read-Only Token Blocked on All Write Endpoints
		writeRoutes := []struct {
			method string
			path   string
		}{
			{"POST", "/api/v1/org/rbac-org/rules"},
			{"DELETE", "/api/v1/org/rbac-org/rules/123"},
			{"POST", "/api/v1/org/rbac-org/insurance/claims"},
			{"POST", "/api/v1/org/rbac-org/zombies/pr"},
			{"POST", "/api/v1/org/rbac-org/partners"},
		}

		for _, route := range writeRoutes {
			req, _ := http.NewRequest(route.method, p8ApiURL+route.path, bytes.NewReader([]byte("{}")))
			req.Header.Set("Authorization", "Bearer "+viewerJWT)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			resp.Body.Close()
			assert.Equal(t, http.StatusForbidden, resp.StatusCode, "Read-Only role should get 403 Forbidden for "+route.method+" "+route.path)
		}

		// Scenario 2: Admin Token Allowed on All Write Endpoints
		// Note: The global test pool and ctx are already set up at the start of TestPhase8SystemE2E
		_, err := pool.Exec(ctx, "INSERT INTO organizations (github_installation_id, github_org_name) VALUES (888, 'rbac-org') ON CONFLICT DO NOTHING")
		require.NoError(t, err)

		var rbacOrgID string
		err = pool.QueryRow(ctx, "SELECT id FROM organizations WHERE github_org_name = 'rbac-org'").Scan(&rbacOrgID)
		if err == nil {
			_, _ = pool.Exec(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name) VALUES ($1, 333, 'rbac-repo', 'rbac-org/rbac-repo') ON CONFLICT DO NOTHING", rbacOrgID)
		}

		for _, route := range writeRoutes {
			req, _ := http.NewRequest(route.method, p8ApiURL+route.path, bytes.NewReader([]byte("{}")))
			req.Header.Set("Authorization", "Bearer "+adminJWT)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			resp.Body.Close()
			assert.NotEqual(t, http.StatusForbidden, resp.StatusCode, "Admin role should not get 403 for "+route.method+" "+route.path)
			assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode, "Admin role should not get 401 for "+route.method+" "+route.path)
		}

		// Scenario 3: Service Token Allowed on Service-Token Routes
		serviceRoutes := []struct {
			method string
			path   string
		}{
			{"POST", "/api/v1/sync"},
			{"POST", "/api/v1/diff"},
			{"POST", "/api/v1/otel/webhook"},
		}

		for _, route := range serviceRoutes {
			req, _ := http.NewRequest(route.method, p8ApiURL+route.path, bytes.NewReader([]byte("{}")))
			req.Header.Set("Authorization", "Bearer "+p8RegistryAPIToken)
			req.Header.Set("Content-Type", "application/json")

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			resp.Body.Close()
			assert.NotEqual(t, http.StatusForbidden, resp.StatusCode, "Service token should not get 403 for "+route.method+" "+route.path)
			assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode, "Service token should not get 401 for "+route.method+" "+route.path)

			// User JWT should fail on service token routes
			req, _ = http.NewRequest(route.method, p8ApiURL+route.path, bytes.NewReader([]byte("{}")))
			req.Header.Set("Authorization", "Bearer "+adminJWT)
			req.Header.Set("Content-Type", "application/json")

			resp, err = http.DefaultClient.Do(req)
			require.NoError(t, err)
			resp.Body.Close()
			assert.True(t, resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized, "User JWT should be denied for service route")
		}

		// Scenario 4: No Token Returns 401
		reqNoToken, _ := http.NewRequest("POST", p8ApiURL+"/api/v1/org/rbac-org/rules", bytes.NewReader([]byte("{}")))
		respNoToken, err := http.DefaultClient.Do(reqNoToken)
		require.NoError(t, err)
		respNoToken.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, respNoToken.StatusCode, "Missing token should get 401 Unauthorized")
	})
}
