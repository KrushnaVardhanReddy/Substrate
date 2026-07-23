package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"e2e/helpers"
)

const (
	p10RegistryAPIToken = "local-dev-token"
	p10JWTSecret        = "local-jwt-secret"
	p10ApiURL           = "http://localhost:8090"
	p10DbURL            = "postgres://postgres:postgres@localhost:5432/substrate?sslmode=disable"
)

func waitForP10Services(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(p10ApiURL + "/health")
	if err != nil || resp.StatusCode != 200 {
		t.Skipf("API server not reachable at %s. Skipping E2E test.", p10ApiURL)
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func setupP10Database(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, p10DbURL)
	if err != nil {
		t.Skip("Failed to connect to real PostgreSQL. Skipping E2E test.")
	}

	_, err = pool.Exec(ctx, "DELETE FROM endpoint_traffic")
	if err != nil {
		t.Logf("Failed to delete endpoint_traffic: %v", err)
	}
	_, err = pool.Exec(ctx, "DELETE FROM governance_rules")
	if err != nil {
		t.Logf("Failed to delete governance_rules: %v", err)
	}
	_, err = pool.Exec(ctx, "DELETE FROM organizations")
	if err != nil {
		t.Logf("Failed to delete organizations: %v", err)
	}

	return pool
}

func createP10JWT(org, role string) string {
	claims := jwt.MapClaims{
		"orgs": map[string]interface{}{
			org: role,
		},
		"exp": time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(p10JWTSecret))
	return tokenString
}

func TestPhase10SystemE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	waitForP10Services(t)

	ctx := context.Background()
	pool := setupP10Database(t)
	defer pool.Close()

	var orgID string
	err := pool.QueryRow(ctx, "INSERT INTO organizations (github_installation_id, github_org_name) VALUES (1010, 'phase10-org') RETURNING id").Scan(&orgID)
	require.NoError(t, err)

	repoID := "00000000-0000-0000-0000-000000000001"
	_, err = pool.Exec(ctx, "INSERT INTO repositories (id, org_id, github_repo_id, name, full_name) VALUES ($1, $2, 1234, 'billing-api', 'phase10-org/billing-api')", repoID, orgID)
	require.NoError(t, err)

	t.Run("Scenario 1: OTel webhook ingestion", func(t *testing.T) {
		spans := []map[string]interface{}{
			{
				"repo_id":   repoID,
				"method":    "GET",
				"path":      "/api/v1/users",
				"timestamp": time.Now().Format(time.RFC3339),
			},
		}

		body, _ := json.Marshal(spans)
		req, err := http.NewRequest("POST", p10ApiURL+"/api/v1/otel/webhook", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+p10RegistryAPIToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	})

	t.Run("Scenario 2: Zombie detection query", func(t *testing.T) {
		req, err := http.NewRequest("GET", p10ApiURL+"/api/v1/org/phase10-org/zombies", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+p10RegistryAPIToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var zombies []map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&zombies)
		// Should succeed, we just ensure the endpoint returns a valid response
	})

	t.Run("Scenario 3: Governance rules CRUD and API Governance Comment", func(t *testing.T) {
		rulePayload := map[string]interface{}{
			"rule_text": "All endpoints must have an X-Correlation-ID header",
		}
		body, _ := json.Marshal(rulePayload)

		req, err := http.NewRequest("POST", p10ApiURL+"/api/v1/org/phase10-org/rules", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+p10RegistryAPIToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		// 1. Create a real draft PR in local forgejo to act as our target for comments
		prURL, err := helpers.CreateDraftPR(ctx, "testorg", "testrepo", "phase10-e2e-branch-"+fmt.Sprint(time.Now().Unix()), "test patch", "test PR for governance", "body")
		if err != nil {
			t.Skipf("Failed to create draft PR in local forgejo, skipping comment assertion: %v", err)
		}

		// The mock prURL is typically "http://localhost:3000/testorg/testrepo/pulls/1" or similar
		// We'll extract the issue number. Forgejo PRs are also issues.
		// For simplicity, we just use 1 if it's the first PR, or parse it.
		// The CreateDraftPR in phase8 uses issue 1 if not parsed properly.
		// Actually, let's just use 1 as a fallback, but the API creates them sequentially.
		// A cleaner way is to parse the ID from the URL, but local Forgejo might return it reliably.
		var prNumber int = 1

		// Try to parse PR number if it exists at the end of the URL
		var urlParts []string
		for i := len(prURL) - 1; i >= 0; i-- {
			if prURL[i] == '/' {
				urlParts = append(urlParts, prURL[i+1:])
				break
			}
		}
		if len(urlParts) > 0 && len(urlParts[0]) > 0 {
			var id int
			_, _ = fmt.Sscanf(urlParts[0], "%d", &id)
			if id > 0 {
				prNumber = id
			}
		}

		// 2. Post a diff report and expect a PR comment containing "API Governance"
		diffReq := map[string]interface{}{
			"diff_report":         json.RawMessage(`{"summary":{"breaking_count":1},"breaking_changes":[{"description":"Breaking change detected"}]}`),
			"is_audit_mode":       false,
			"org":                 "phase10-org",
			"provider_repo":       "testorg/testrepo",
			"pr_number":           prNumber,
			"commit_sha":          "abc123sha",
			"head_schema_content": "openapi: 3.0.0\ninfo:\n  version: 2.0.0", // Lacks x-correlation-id
			"schema_type":         "openapi",
		}
		diffBody, _ := json.Marshal(diffReq)

		reqDiff, err := http.NewRequest("POST", p10ApiURL+"/api/v1/diff", bytes.NewReader(diffBody))
		require.NoError(t, err)
		reqDiff.Header.Set("Content-Type", "application/json")
		reqDiff.Header.Set("Authorization", "Bearer "+p10RegistryAPIToken)

		respDiff, err := http.DefaultClient.Do(reqDiff)
		require.NoError(t, err)
		defer respDiff.Body.Close()

		assert.Equal(t, http.StatusCreated, respDiff.StatusCode)

		// 3. Assert the PR comment was created on the local forgejo instance
		// We must wait a brief moment for async workers or API calls to settle (though diff handler does it synchronously for PR comments)
		time.Sleep(1 * time.Second)
		comments, err := helpers.ListIssueComments(ctx, "testorg", "testrepo", prNumber)
		require.NoError(t, err)

		foundGovernanceComment := false
		for _, c := range comments {
			if strings.Contains(c, "API Governance") {
				foundGovernanceComment = true
				break
			}
		}
		assert.True(t, foundGovernanceComment, "Expected to find a PR comment containing 'API Governance'")
	})

	t.Run("Scenario 4: Auto-SDK generation trigger", func(t *testing.T) {
		pushPayload := map[string]interface{}{
			"installation_id": 1010,
			"org":             "phase10-org",
			"repo":            "phase10-org/billing-api",
			"commit_sha":      "commitsha123",
			"branch":          "main",
			"files": []map[string]interface{}{
				{
					"path":    "openapi.yaml",
					"content": "openapi: 3.0.0\ninfo:\n  title: Billing API\n  version: 2.0.0\n",
				},
			},
		}
		body, _ := json.Marshal(pushPayload)

		req, err := http.NewRequest("POST", p10ApiURL+"/api/v1/webhook", bytes.NewReader(body))
		require.NoError(t, err)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+p10RegistryAPIToken)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	})
}
