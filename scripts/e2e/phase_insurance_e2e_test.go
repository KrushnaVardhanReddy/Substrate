package main

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func setupInsuranceDatabase(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	require.NoError(t, err, "Failed to connect to real PostgreSQL")

	_, err = pool.Exec(ctx, "DELETE FROM insurance_claims")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM insurance_policies")
	require.NoError(t, err)
	_, err = pool.Exec(ctx, "DELETE FROM organizations")
	require.NoError(t, err)

	return pool
}

func TestInsuranceE2E(t *testing.T) {
	waitForServices(t)
	pool := setupInsuranceDatabase(t)
	defer pool.Close()

	t.Run("Get Policy", func(t *testing.T) {
		ctx := context.Background()

		// Clean up before test
		_, err := pool.Exec(ctx, "DELETE FROM insurance_claims")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, "DELETE FROM insurance_policies")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, "DELETE FROM organizations")
		require.NoError(t, err)

		orgID := "11111111-1111-1111-1111-111111111111"
		_, err = pool.Exec(ctx, "INSERT INTO organizations (id, github_installation_id, github_org_name) VALUES ($1, $2, $3)", orgID, 1001, "ins-org-1")
		require.NoError(t, err)

		policyLimit := 5000000 // $50,000.00
		_, err = pool.Exec(ctx, "INSERT INTO insurance_policies (org_id, policy_limit_cents, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())", orgID, policyLimit)
		require.NoError(t, err)

		req, err := http.NewRequest("GET", apiURL+"/api/v1/org/ins-org-1/insurance/policy", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+registryAPIToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		require.Equal(t, float64(policyLimit), result["policy_limit_cents"])
	})

	t.Run("List Claims", func(t *testing.T) {
		ctx := context.Background()

		// Clean up before test
		_, err := pool.Exec(ctx, "DELETE FROM insurance_claims")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, "DELETE FROM insurance_policies")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, "DELETE FROM organizations")
		require.NoError(t, err)

		orgID := "22222222-2222-2222-2222-222222222222"
		_, err = pool.Exec(ctx, "INSERT INTO organizations (id, github_installation_id, github_org_name) VALUES ($1, $2, $3)", orgID, 1002, "ins-org-2")
		require.NoError(t, err)

		policyID := "33333333-3333-3333-3333-333333333333"
		policyLimit := 5000000 // $50,000.00
		_, err = pool.Exec(ctx, "INSERT INTO insurance_policies (id, org_id, policy_limit_cents, created_at, updated_at) VALUES ($1, $2, $3, NOW(), NOW())", policyID, orgID, policyLimit)
		require.NoError(t, err)

		claimID1 := "44444444-4444-4444-4444-444444444444"
		claimID2 := "55555555-5555-5555-5555-555555555555"

		_, err = pool.Exec(ctx, `
			INSERT INTO insurance_claims (id, org_id, policy_id, github_pr_url, incident_date, status, amount_cents, created_at, updated_at)
			VALUES
			($1, $2, $3, 'https://github.com/acme/repo/pull/1', NOW() - INTERVAL '1 day', 'PENDING', 50000, NOW() - INTERVAL '1 day', NOW() - INTERVAL '1 day'),
			($4, $2, $3, 'https://github.com/acme/repo/pull/2', NOW(), 'APPROVED', 25000, NOW(), NOW())
		`, claimID1, orgID, policyID, claimID2)
		require.NoError(t, err)

		req, err := http.NewRequest("GET", apiURL+"/api/v1/org/ins-org-2/insurance/claims", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+registryAPIToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusOK, resp.StatusCode)

		var result []map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		require.Len(t, result, 2)

		// The query in pgstore orders by created_at DESC, so the second insert (claimID2) should be first
		require.Equal(t, claimID2, result[0]["id"])
		require.Equal(t, "APPROVED", result[0]["status"])
		require.Equal(t, float64(25000), result[0]["amount_cents"])

		require.Equal(t, claimID1, result[1]["id"])
		require.Equal(t, "PENDING", result[1]["status"])
		require.Equal(t, float64(50000), result[1]["amount_cents"])
	})

	t.Run("Not Found", func(t *testing.T) {
		ctx := context.Background()

		// Clean up before test
		_, err := pool.Exec(ctx, "DELETE FROM insurance_claims")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, "DELETE FROM insurance_policies")
		require.NoError(t, err)
		_, err = pool.Exec(ctx, "DELETE FROM organizations")
		require.NoError(t, err)

		orgID := "66666666-6666-6666-6666-666666666666"
		_, err = pool.Exec(ctx, "INSERT INTO organizations (id, github_installation_id, github_org_name) VALUES ($1, $2, $3)", orgID, 1003, "ins-org-no-policy")
		require.NoError(t, err)

		req, err := http.NewRequest("GET", apiURL+"/api/v1/org/ins-org-no-policy/insurance/policy", nil)
		require.NoError(t, err)
		req.Header.Set("Authorization", "Bearer "+registryAPIToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		require.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}
