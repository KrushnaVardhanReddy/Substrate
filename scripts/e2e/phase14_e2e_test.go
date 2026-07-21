package main

import (
	"bytes"
	"context"
	"encoding/json"
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
	p14RegistryAPIToken = "local-dev-token"
	p14ApiURL           = "http://localhost:8090"
	p14DbURL            = "postgres://postgres:postgres@localhost:5432/substrate?sslmode=disable"
)

func waitForP14Services(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 5; i++ {
		resp, err := client.Get(p14ApiURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("API server not reachable at %s. Please ensure 'make api' and 'make postgres' are running.", p14ApiURL)
}

func setupP14Database(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, p14DbURL)
	require.NoError(t, err, "Failed to connect to real PostgreSQL")

	// Clean up tables relevant to Phase 14
	_, err = pool.Exec(ctx, "DELETE FROM repositories")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM organizations")
	require.NoError(t, err)

	return pool
}

func TestPhase14SystemE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	ctx := context.Background()

	waitForP14Services(t)

	// Build the CLI binary
	engineDir, err := filepath.Abs("../../engine")
	require.NoError(t, err)
	binPath := filepath.Join(engineDir, "substrate_p14_test_bin")
	cmdBuild := exec.Command("go", "build", "-o", "substrate_p14_test_bin", "./cmd/substrate")
	cmdBuild.Dir = engineDir
	require.NoError(t, cmdBuild.Run(), "Failed to compile the substrate CLI")
	defer os.Remove(binPath) // Cleanup

	// Database operations for setup
	pool := setupP14Database(t)
	defer pool.Close()

	var orgID string
	err = pool.QueryRow(ctx, "INSERT INTO organizations (github_installation_id, github_org_name) VALUES (999, 'p14org') RETURNING id").Scan(&orgID)
	require.NoError(t, err)

	var repoID string
	err = pool.QueryRow(ctx, "INSERT INTO repositories (org_id, github_repo_id, name, full_name) VALUES ($1, 111, 'p14repo', 'p14org/p14repo') RETURNING id", orgID).Scan(&repoID)
	require.NoError(t, err)

	t.Run("Scenario 1: Contract Score Badge (P14-T04)", func(t *testing.T) {
		resp, err := http.Get(p14ApiURL + "/api/badges/p14org/p14repo")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, "image/svg+xml", resp.Header.Get("Content-Type"))
	})

	t.Run("Scenario 2: Archaeology CLI (P14-T05)", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Initialize git repository
		cmdGitInit := exec.Command("git", "init")
		cmdGitInit.Dir = tmpDir
		require.NoError(t, cmdGitInit.Run())

		cmdGitConfigUser := exec.Command("git", "config", "user.name", "Test User")
		cmdGitConfigUser.Dir = tmpDir
		require.NoError(t, cmdGitConfigUser.Run())

		cmdGitConfigEmail := exec.Command("git", "config", "user.email", "test@example.com")
		cmdGitConfigEmail.Dir = tmpDir
		require.NoError(t, cmdGitConfigEmail.Run())

		// Write a dummy spec
		err := os.WriteFile(filepath.Join(tmpDir, "openapi.yaml"), []byte("openapi: 3.0.0\ninfo:\n  version: 1.0.0\n"), 0644)
		require.NoError(t, err)

		cmdGitAdd := exec.Command("git", "add", "openapi.yaml")
		cmdGitAdd.Dir = tmpDir
		require.NoError(t, cmdGitAdd.Run())

		cmdGitCommit := exec.Command("git", "commit", "-m", "initial commit")
		cmdGitCommit.Dir = tmpDir
		require.NoError(t, cmdGitCommit.Run())

		// Run archaeology binary
		cmd := exec.Command(binPath, "archaeology", "--repo", tmpDir, "--since", "30-days")
		var outBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &outBuf

		err = cmd.Run()
		require.NoError(t, err, "Archaeology command failed: %s", outBuf.String())

		assert.Contains(t, outBuf.String(), "Running archaeology scan for the past")
	})

	t.Run("Scenario 3: Contract Negotiation (P14-T01)", func(t *testing.T) {
		payload := map[string]interface{}{
			"action": "created",
			"comment": map[string]interface{}{
				"id":   123,
				"body": "This is a comment",
			},
			"reaction": map[string]interface{}{
				"content": "+1",
			},
			"sender": map[string]interface{}{
				"login": "alice",
			},
			"repository": map[string]interface{}{
				"name": "p14repo",
				"owner": map[string]interface{}{
					"login": "p14org",
				},
			},
			"issue": map[string]interface{}{
				"number": 1,
			},
		}

		body, err := json.Marshal(payload)
		require.NoError(t, err)

		req, err := http.NewRequest("POST", p14ApiURL+"/api/v1/webhook/reaction", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+p14RegistryAPIToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	})
}
