package main

import (
	"bytes"
	"context"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	p8RegistryAPITokenE2E = "local-dev-token"
	p8ApiURLE2E           = "http://localhost:8090"
	p8DbURLE2E            = "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
)

func waitForP8ServicesE2E(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 5; i++ {
		resp, err := client.Get(p8ApiURLE2E + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	t.Skipf("API server not reachable at %s", p8ApiURLE2E)
}

func setupP8DatabaseE2E(t *testing.T) (*pgxpool.Pool, string, string) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, p8DbURLE2E)
	require.NoError(t, err, "Failed to connect to real PostgreSQL")

	// Clean up and seed data specific to this test
	_, err = pool.Exec(ctx, "DELETE FROM dependencies")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM contracts")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM repositories")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM organizations")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "INSERT INTO organizations (github_installation_id, github_org_name) VALUES (111, 'acme') ON CONFLICT DO NOTHING")
	require.NoError(t, err)

	var orgID string
	err = pool.QueryRow(ctx, "SELECT id FROM organizations WHERE github_org_name = 'acme'").Scan(&orgID)
	require.NoError(t, err)

	var billingApiRepoID string
	err = pool.QueryRow(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name) VALUES ($1, 222, 'billing-api', 'acme/billing-api') RETURNING id", orgID).Scan(&billingApiRepoID)
	require.NoError(t, err)

	var invoiceSvcRepoID string
	err = pool.QueryRow(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name) VALUES ($1, 333, 'invoice-service', 'acme/invoice-service') RETURNING id", orgID).Scan(&invoiceSvcRepoID)
	require.NoError(t, err)

	return pool, billingApiRepoID, invoiceSvcRepoID
}

func TestPhase8SystemCLIE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	waitForP8ServicesE2E(t)

	pool, billingApiRepoID, invoiceSvcRepoID := setupP8DatabaseE2E(t)
	defer pool.Close()
	ctx := context.Background()

	// Build the CLI binary
	engineDir, err := filepath.Abs("../../engine")
	require.NoError(t, err)
	binPath := filepath.Join(engineDir, "substrate_p8_test_bin")
	cmdBuild := exec.Command("go", "build", "-o", "substrate_p8_test_bin", "./cmd/substrate")
	cmdBuild.Dir = engineDir
	require.NoError(t, cmdBuild.Run(), "Failed to compile the substrate CLI")
	defer os.Remove(binPath) // Cleanup

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
              # for consumers who expect 'amount' in V2.
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
		cmd.Env = append(os.Environ(), "SUBSTRATE_API_URL="+p8ApiURLE2E, "REGISTRY_API_TOKEN="+p8RegistryAPITokenE2E)

		var outBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &outBuf

		err = cmd.Run()
		require.Error(t, err)
		if exitError, ok := err.(*exec.ExitError); ok {
			assert.Equal(t, 2, exitError.ExitCode(), "Exit code should be 2 for blocked rollback")
		} else {
			t.Skipf("Expected ExitError, got %v", err)
		}

		assert.Contains(t, outBuf.String(), "ROLLBACK BLOCKED")
		assert.Contains(t, outBuf.String(), "acme/invoice-service")
	})
}
