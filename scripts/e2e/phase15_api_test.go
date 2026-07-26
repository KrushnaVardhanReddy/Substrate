package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	p15ApiURL    = "http://localhost:8090"
	p15DbURL     = "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
	p15JWTSecret = "local-jwt-secret"
)

func createP15JWT(org, role string) string {
	hash := sha256.Sum256([]byte(p15JWTSecret))
	key, _ := paseto.V4SymmetricKeyFromBytes(hash[:])

	token := paseto.NewToken()
	token.Set("orgs", map[string]string{org: role})
	token.SetExpiration(time.Now().Add(time.Hour))

	return token.V4Encrypt(key, nil)
}

func waitForP15Services(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 5; i++ {
		resp, err := client.Get(p15ApiURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	t.Skipf("API server not reachable at %s. Please ensure it is running.", p15ApiURL)
}

func setupP15Database(t *testing.T) (*pgxpool.Pool, string, string) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, p15DbURL)
	require.NoError(t, err, "Failed to connect to real PostgreSQL")

	// Clean up tables relevant to Phase 15
	_, err = pool.Exec(ctx, "DELETE FROM insurance_claims")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM insurance_policies")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM organizations WHERE github_org_name = 'mcp-org'")
	require.NoError(t, err)

	// Create org (with is_oss = true)
	var orgID string
	err = pool.QueryRow(ctx, "INSERT INTO organizations (github_installation_id, github_org_name, is_oss) VALUES (1515, 'mcp-org', true) ON CONFLICT (github_installation_id) DO UPDATE SET is_oss = EXCLUDED.is_oss RETURNING id").Scan(&orgID)
	require.NoError(t, err)

	// Create insurance policy
	var policyID string
	err = pool.QueryRow(ctx, "INSERT INTO insurance_policies (org_id, policy_limit_cents) VALUES ($1, 5000000) RETURNING id", orgID).Scan(&policyID)
	require.NoError(t, err)

	// Create existing incident in insurance_claims
	_, err = pool.Exec(ctx, "INSERT INTO insurance_claims (id, org_id, policy_id, github_pr_url, incident_date, status, amount_cents) VALUES ($1, $2, $3, $4, NOW(), $5, $6)", uuid.New().String(), orgID, policyID, "https://github.com/mcp-org/repo/pull/10", "APPROVED", 15000)
	require.NoError(t, err)

	return pool, orgID, policyID
}

func TestPhase15API_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	waitForP15Services(t)

	// Build the CLI binary
	engineDir, err := filepath.Abs("../../engine")
	require.NoError(t, err)
	binPath := filepath.Join(engineDir, "substrate_p15_test_bin")
	cmdBuild := exec.Command("go", "build", "-o", "substrate_p15_test_bin", "./cmd/substrate")
	cmdBuild.Dir = engineDir
	require.NoError(t, cmdBuild.Run(), "Failed to compile the substrate CLI")
	defer os.Remove(binPath) // Cleanup

	// Database operations for setup
	pool, _, _ := setupP15Database(t)
	defer pool.Close()

	t.Run("Scenario 1: NL Governance Rule generation", func(t *testing.T) {
		adminJWT := createP15JWT("mcp-org", "admin")

		reqBody := map[string]interface{}{
			"prompt": "All payment APIs must require authentication",
		}
		body, _ := json.Marshal(reqBody)

		req, err := http.NewRequest("POST", p15ApiURL+"/api/governance/generate-cel", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminJWT)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)
		var respBody map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&respBody)
		require.NoError(t, err)

		assert.NotEmpty(t, respBody["cel"])
	})

	t.Run("Scenario 2: Marketplace Plugin Install", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "e2e_marketplace_")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		cfgPath := filepath.Join(tempDir, "substrate.yaml")
		err = os.WriteFile(cfgPath, []byte("service: mock-service\nschema_type: openapi\n"), 0644)
		require.NoError(t, err)

		cmd := exec.Command(binPath, "plugin", "install", "substrate-plugin-hipaa")
		cmd.Dir = tempDir
		cmd.Env = append(os.Environ(), "SUBSTRATE_API_URL="+p15ApiURL)

		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "Plugin install failed. Output: %s", string(out))
		assert.Contains(t, string(out), "Plugin 'substrate-plugin-hipaa' installed successfully")
	})

	t.Run("Scenario 4: Schema Insurance Claims", func(t *testing.T) {
		adminJWT := createP15JWT("mcp-org", "admin")

		claimReq := map[string]interface{}{
			"github_pr_url": "https://github.com/mcp-org/repo/pull/42",
			"incident_date": time.Now().Format(time.RFC3339),
			"amount_cents":  5000,
		}
		body, _ := json.Marshal(claimReq)

		req, err := http.NewRequest("POST", p15ApiURL+"/api/v1/org/mcp-org/insurance/claims", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+adminJWT)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var claim map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&claim)
		require.NoError(t, err)

		require.NotEmpty(t, claim)
		assert.Equal(t, "https://github.com/mcp-org/repo/pull/42", claim["github_pr_url"])
		assert.Equal(t, "PENDING", claim["status"])
	})
}
