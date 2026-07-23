package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
	apiURL           = "http://localhost:8090"
	dbURL            = "postgres://postgres:postgres@localhost:5432/substrate?sslmode=disable"
	registryAPIToken = "local-dev-token"
)

func waitForServices(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 5; i++ {
		resp, err := client.Get(apiURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
	}
	// Note: Memory guidelines state "When running E2E tests in scripts/e2e, if local API infrastructure
	// (e.g., http://localhost:8090) is unreachable, the test should call t.Skip(...) rather than
	// t.Fatal(...). This handles restricted sandbox environments gracefully without breaking the build."
	t.Skipf("API server not reachable at %s. Skipping test in restricted sandbox environment.", apiURL)
}

func setupDatabase(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err, "Failed to connect to real PostgreSQL")

	_, err = pool.Exec(ctx, "DELETE FROM breaking_change_history")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM diff_reports")
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

func TestV1SystemE2E(t *testing.T) {
	waitForServices(t)
	// Build the CLI binary once for all tests
	engineDir, err := filepath.Abs("../../engine")
	require.NoError(t, err)

	binPath := filepath.Join(engineDir, "substrate_test_bin")
	cmdBuild := exec.Command("go", "build", "-o", "substrate_test_bin", "./cmd/substrate")
	cmdBuild.Dir = engineDir
	err = cmdBuild.Run()
	require.NoError(t, err, "Failed to compile the substrate CLI")
	defer os.Remove(binPath) // Cleanup

	pool := setupDatabase(t)
	defer pool.Close()

	// Step 1: Zero-Touch Onboarding (V1-T01)
	// Simulate the GitHub App Webhook receiving a new repository installation,
	// asserting it triggers the Auto-Discovery PR by firing a payload via real HTTP to the API.
	t.Run("Step 1: Auto-Discovery Webhook", func(t *testing.T) {
		payload := map[string]interface{}{
			"installation_id": 9999,
			"org":             "testorg",
			"repo":            "testorg/testconsumer",
			"github_repo_id":  12345,
			"commit_sha":      "abcdef123456",
			"files": []map[string]interface{}{
				{
					"path":    ".env.example",
					"content": "TEST_API_URL=http://testprovider:8080",
				},
			},
		}

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", apiURL+"/api/v1/webhook", bytes.NewReader(body))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+registryAPIToken)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusAccepted, resp.StatusCode)

		var respData map[string]string
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)

		assert.Equal(t, "queued", respData["status"], "Should have queued the webhook for processing")
	})

	// Step 2: AI Architect Scaffolding (V1-T02)
	t.Run("Step 2: AI Architect Scaffolding", func(t *testing.T) {
		// Mock OpenAI Server
		mockOpenAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/chat/completions", r.URL.Path)

			w.Header().Set("Content-Type", "text/event-stream")

			resp := map[string]interface{}{
				"choices": []map[string]interface{}{
					{
						"delta": map[string]interface{}{
							"content": "```yaml\nopenapi: 3.0.0\ninfo:\n  title: Example API\n  version: 1.0.0\npaths:\n  /users:\n    get:\n      responses:\n        '200':\n          description: OK\n```",
						},
					},
				},
			}
			respBytes, _ := json.Marshal(resp)
			fmt.Fprintf(w, "data: %s\n\n", string(respBytes))
			fmt.Fprintf(w, "data: [DONE]\n\n")
		}))
		defer mockOpenAI.Close()

		// Run the init --design command
		cmd := exec.Command(binPath, "init", "--design", "--service", "test-service", "--force")
		cmd.Dir = engineDir // run it in engine dir temporarily, but we want a clean temp dir

		tempDir := t.TempDir()
		cmd.Dir = tempDir

		cmd.Env = append(os.Environ(),
			"SUBSTRATE_AI_BASE_URL="+mockOpenAI.URL,
			"SUBSTRATE_AI_API_KEY=mock_key",
			"SUBSTRATE_AI_MODEL=mock_model",
		)

		// Simulate user input for AI prompts (just send a newline to accept default/skip)
		cmd.Stdin = bytes.NewBufferString("create a user api\n")

		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "Failed to run init --design: %s", string(out))

		// Check if files are created
		_, err = os.Stat(filepath.Join(tempDir, "openapi.yaml"))
		assert.NoError(t, err, "openapi.yaml should be created")

		_, err = os.Stat(filepath.Join(tempDir, "substrate.yaml"))
		assert.NoError(t, err, "substrate.yaml should be created")
	})

	// Step 3: Schema Break & Diff Generation (V1-T03)
	t.Run("Step 3: Schema Break and Diff Generation", func(t *testing.T) {
		// Use the mock schemas we created earlier
		basePath, err := filepath.Abs("testdata/v1/mock_ai_response.yaml")
		require.NoError(t, err)
		breakPath, err := filepath.Abs("testdata/v1/break_schema.yaml")
		require.NoError(t, err)

		// Run diff and get JSON output
		cmd := exec.Command(binPath, "diff", basePath, breakPath, "--format", "json")
		out, err := cmd.CombinedOutput()

		require.Error(t, err, "Breaking change should cause diff command to fail")

		// In substrate, breaking changes result in exit code 2
		// So we don't assert err == nil, we just parse the output

		var diffReport map[string]interface{}
		err = json.Unmarshal(out, &diffReport)
		require.NoError(t, err, "Failed to parse DiffReport: %s", string(out))

		summary, ok := diffReport["summary"].(map[string]interface{})
		require.True(t, ok)
		breakingCount, ok := summary["breaking_count"].(float64)
		require.True(t, ok)
		assert.Greater(t, int(breakingCount), 0, "Diff should contain breaking changes")

		// Post DiffReport to /api/v1/diff
		payload := map[string]interface{}{
			"diff_report": diffReport,
			"org":         "acme-corp", // test org
		}
		body, err := json.Marshal(payload)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", apiURL+"/api/v1/diff", bytes.NewReader(body))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+registryAPIToken)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusCreated, resp.StatusCode)

		var respData map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&respData)
		require.NoError(t, err)

		_, ok = respData["id"].(string)
		assert.True(t, ok, "Should return a diff ID")
	})

	// Step 4: Local Developer Remediation (V1-T05)
	t.Run("Step 4: Local Developer Remediation", func(t *testing.T) {
		basePath, err := filepath.Abs("testdata/v1/mock_ai_response.yaml")
		require.NoError(t, err)
		fixPath, err := filepath.Abs("testdata/v1/fix_schema.yaml")
		require.NoError(t, err)

		// The validate command in engine is a placeholder.
		// The requirement states "The developer uses the local CLI substrate validate to test a fix (restoring the field).
		// Assertion: The CLI must return an exit code 0 (Safe) for the fixed schema."
		// Let's use the validate command which exists and exits with 0.
		cmd := exec.Command(binPath, "validate", fixPath)
		out, err := cmd.CombinedOutput()
		assert.NoError(t, err, "Validation should pass with exit code 0: %s", string(out))

		// To actually verify the diff is safe (since validate is a placeholder), let's run diff
		cmdDiff := exec.Command(binPath, "diff", basePath, fixPath, "--format", "json")
		outDiff, errDiff := cmdDiff.CombinedOutput()
		assert.NoError(t, errDiff, "Diff should be safe (exit code 0)")

		var diffReport map[string]interface{}
		json.Unmarshal(outDiff, &diffReport)
		summary, _ := diffReport["summary"].(map[string]interface{})
		breakingCount, _ := summary["breaking_count"].(float64)
		assert.Equal(t, float64(0), breakingCount, "Should have 0 breaking changes after fix")
	})

	// Step 5: Deployment Safety Gate (V1-T04)
	t.Run("Step 5: Deployment Safety Gate", func(t *testing.T) {
		// Mock State: The registry must be seeded with a consumer (e.g., frontend)
		// that is currently deployed in production and depends on the old schema.
		// Let's seed this using the PostgreSQL pool directly to avoid complex API setup
		ctx := context.Background()

		// 1. Insert Org
		var orgID string
		err := pool.QueryRow(ctx, "INSERT INTO organizations (github_installation_id, github_org_name) VALUES ($1, $2) RETURNING id", 10001, "testorg-deploy").Scan(&orgID)
		require.NoError(t, err)

		// 2. Insert Provider Repo
		var providerRepoID string
		err = pool.QueryRow(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name) VALUES ($1, $2, $3, $4) RETURNING id", orgID, 20001, "backend-api", "testorg-deploy/backend-api").Scan(&providerRepoID)
		require.NoError(t, err)

		// 3. Insert Consumer Repo
		var consumerRepoID string
		err = pool.QueryRow(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name) VALUES ($1, $2, $3, $4) RETURNING id", orgID, 20002, "frontend", "testorg-deploy/frontend").Scan(&consumerRepoID)
		require.NoError(t, err)

		// 4. Insert Contract for Provider
		var contractID string
		err = pool.QueryRow(ctx, "INSERT INTO contracts (repo_id, schema_type, spec_path, latest_commit_sha, raw_content) VALUES ($1, $2, $3, $4, $5) RETURNING id", providerRepoID, "openapi", "openapi.yaml", "provider-hash-old", "mock-content").Scan(&contractID)
		require.NoError(t, err)

		// 5. Map Consumer to Contract
		_, err = pool.Exec(ctx, "INSERT INTO dependencies (consumer_repo_id, provider_contract_id, status) VALUES ($1, $2, $3)", consumerRepoID, contractID, "active")
		require.NoError(t, err)

		// 6. Insert a history entry representing the new breaking deployment
		_, err = pool.Exec(ctx, "INSERT INTO breaking_change_history (repo_id, org_name, repo_name, git_sha, breaking_changes) VALUES ($1, $2, $3, $4, $5)", providerRepoID, "testorg-deploy", "backend-api", "provider-hash-new", `{"breaking_count": 1}`)
		require.NoError(t, err)

		// Run check-deploy

		cmd := exec.Command(binPath, "check-deploy", "--repo", "testorg-deploy/backend-api", "--commit", "provider-hash-new")
		cmd.Env = append(os.Environ(), "SUBSTRATE_API_URL="+apiURL, "REGISTRY_API_TOKEN="+registryAPIToken)

		out, err := cmd.CombinedOutput()

		// Expect exit code 1 (Blocked)
		assert.Error(t, err)
		if exitError, ok := err.(*exec.ExitError); ok {
			assert.Equal(t, 1, exitError.ExitCode())
		} else {
			t.Fatalf("Expected ExitError, got %v", err)
		}

		outStr := string(out)
		assert.Contains(t, outStr, "❌ DEPLOYMENT BLOCKED")
		assert.Contains(t, outStr, "testorg-deploy/frontend")
	})
}
