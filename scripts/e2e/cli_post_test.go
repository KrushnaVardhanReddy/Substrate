package main

import (

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

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	cliApiURL    = "http://localhost:8090"
	cliDbURL     = "postgres://postgres:postgres@localhost:5432/substrate?sslmode=disable"
	cliJWTSecret = "local-jwt-secret"
)

func createCliJWT(org, role string) string {
	hash := sha256.Sum256([]byte(cliJWTSecret))
	key, _ := paseto.V4SymmetricKeyFromBytes(hash[:])

	token := paseto.NewToken()
	token.Set("orgs", map[string]string{org: role})
	token.SetExpiration(time.Now().Add(time.Hour))

	return token.V4Encrypt(key, nil)
}

func waitForCliServices(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 5; i++ {
		resp, err := client.Get(cliApiURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	t.Skipf("API server not reachable at %s. Gracefully skipping test.", cliApiURL)
}

func setupCliDatabase(t *testing.T) (*pgxpool.Pool, string) {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cliDbURL)
	if err != nil {
		t.Skipf("Failed to connect to postgres: %v", err)
	}

	// Quick ping check
	if err := pool.Ping(ctx); err != nil {
		t.Skipf("Database unreachable: %v", err)
	}

	// Just need organizations table
	_, err = pool.Exec(ctx, "DELETE FROM organizations")
	require.NoError(t, err)

	var orgID string
	err = pool.QueryRow(ctx, "INSERT INTO organizations (github_installation_id, github_org_name) VALUES (1122, 'cli-test-org') RETURNING id").Scan(&orgID)
	require.NoError(t, err)

	return pool, orgID
}

func TestCLIPost(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	waitForCliServices(t)

	// Build the CLI binary
	engineDir, err := filepath.Abs("../../engine")
	require.NoError(t, err)
	binPath := filepath.Join(engineDir, "substrate_cli_post_test_bin")
	cmdBuild := exec.Command("go", "build", "-o", "substrate_cli_post_test_bin", "./cmd/substrate")
	cmdBuild.Dir = engineDir
	require.NoError(t, cmdBuild.Run(), "Failed to compile the substrate CLI")
	defer os.Remove(binPath) // Cleanup

	pool, _ := setupCliDatabase(t)
	if pool != nil {
		defer pool.Close()
	}



	t.Run("Scenario 1: schema-smell", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "e2e_cli_post_smell_")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		schemaPath := filepath.Join(tempDir, "schema.yaml")
		err = os.WriteFile(schemaPath, []byte("openapi: 3.0.0\ninfo:\n  title: Mock API\n"), 0644)
		require.NoError(t, err)

		cmd := exec.Command(binPath, "schema-smell", "--schema", schemaPath)
		cmd.Env = append(os.Environ(),
			"SUBSTRATE_API_URL="+cliApiURL,
			"REGISTRY_API_TOKEN=local-dev-token",
			"SUBSTRATE_AI_BASE_URL=",
			"OPENAI_API_KEY=",
			"SUBSTRATE_AI_API_KEY=",
		)

		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "schema-smell failed. Output: %s", string(out))
		assert.Contains(t, string(out), "FALLBACK_MODE", "Expected fallback mock response for schema-smell without AI keys")
	})

	t.Run("Scenario 2: postmortem", func(t *testing.T) {
		cmd := exec.Command(binPath, "postmortem", "--incident", "2023-10-10T12:00:00Z")
		cmd.Env = append(os.Environ(),
			"SUBSTRATE_API_URL="+cliApiURL,
			"REGISTRY_API_TOKEN=local-dev-token",
		)

		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "postmortem failed. Output: %s", string(out))
		// Our dummy setup doesn't have breaking changes in the prior 24h of 2023-10-10,
		// so postmortem.GeneratePostMortem should output "No breaking changes found in the 24 hours prior to the incident."
		assert.Contains(t, string(out), "No breaking changes found")
	})

	t.Run("Scenario 3: plugin publish", func(t *testing.T) {
		tempDir, err := os.MkdirTemp("", "e2e_cli_post_plugin_")
		require.NoError(t, err)
		defer os.RemoveAll(tempDir)

		pluginPath := filepath.Join(tempDir, "plugin.json")
		pluginJSON := `{
			"name": "test-publish-plugin",
			"description": "test desc",
			"schema": {"rules": [{"id": "r1"}]}
		}`
		err = os.WriteFile(pluginPath, []byte(pluginJSON), 0644)
		require.NoError(t, err)

		cmd := exec.Command(binPath, "plugin", "publish", pluginPath)
		cmd.Env = append(os.Environ(),
			"SUBSTRATE_API_URL="+cliApiURL,
			"REGISTRY_API_TOKEN=local-dev-token",
		)

		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "plugin publish failed. Output: %s", string(out))
		assert.Contains(t, string(out), "Plugin published successfully")

		// Verify plugin published via List API or DB if necessary
		req, _ := http.NewRequest("GET", cliApiURL+"/api/marketplace/plugins", nil)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		var plugins []map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&plugins)

		found := false
		for _, p := range plugins {
			if p["name"] == "test-publish-plugin" {
				found = true
				break
			}
		}
		assert.True(t, found, "Plugin 'test-publish-plugin' should be in the marketplace list")
	})
}
