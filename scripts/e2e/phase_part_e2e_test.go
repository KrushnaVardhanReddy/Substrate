package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	pPartRegistryAPIToken = "local-dev-token"
	pPartJWTSecret        = "local-jwt-secret"
	pPartApiURL           = "http://localhost:8090"
	pPartDbURL            = "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
)

func waitForPPartServices(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(pPartApiURL + "/health")
	if err != nil || resp.StatusCode != 200 {
		t.Skipf("API server not reachable at %s. Skipping E2E test.", pPartApiURL)
	}
	if resp != nil {
		resp.Body.Close()
	}
}

func generatePPartPasetoToken(t *testing.T, org, role string) string {
	hash := sha256.Sum256([]byte(pPartJWTSecret))
	key, err := paseto.V4SymmetricKeyFromBytes(hash[:])
	require.NoError(t, err)

	token := paseto.NewToken()
	token.SetExpiration(time.Now().Add(time.Hour))
	token.Set("orgs", map[string]string{org: role})
	return token.V4Encrypt(key, nil)
}

func setupPPartDatabase(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, pPartDbURL)
	if err != nil {
		t.Skip("Failed to connect to real PostgreSQL. Skipping E2E test.")
	}

	_, err = pool.Exec(ctx, "DELETE FROM partner_integrations")
	require.NoError(t, err)

	return pool
}

func TestPhasePartE2E(t *testing.T) {
	waitForPPartServices(t)
	pool := setupPPartDatabase(t)
	defer pool.Close()

	orgName := "test-org"
	token := generatePPartPasetoToken(t, orgName, "admin")

	client := &http.Client{}

	// 1. Create a partner
	partnerPayload := map[string]string{
		"vendor_name":    "Acme Corp",
		"webhook_url":    "http://example.com/webhook",
		"webhook_secret": "secret123",
	}
	body, err := json.Marshal(partnerPayload)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", pPartApiURL+"/api/v1/org/"+orgName+"/partners", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var createResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&createResp)
	require.NoError(t, err)
	partnerID := createResp["id"].(string)
	assert.NotEmpty(t, partnerID)
	assert.Equal(t, "Acme Corp", createResp["vendor_name"])
	assert.Equal(t, "pending", createResp["status"])

	// 2. List partners
	req, err = http.NewRequest("GET", pPartApiURL+"/api/v1/org/"+orgName+"/partners", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&listResp)
	require.NoError(t, err)
	assert.Len(t, listResp, 1)
	assert.Equal(t, "Acme Corp", listResp[0]["vendor_name"])

	// 3. Update partner
	updatePayload := map[string]string{
		"vendor_name": "Acme Corp Updated",
	}
	updateBody, err := json.Marshal(updatePayload)
	require.NoError(t, err)

	req, err = http.NewRequest("PUT", pPartApiURL+"/api/v1/org/"+orgName+"/partners/"+partnerID, bytes.NewBuffer(updateBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var updateResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&updateResp)
	require.NoError(t, err)
	assert.Equal(t, "Acme Corp Updated", updateResp["vendor_name"])

	// 4. Verify partner (with mock server)
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer mockServer.Close()

	updateUrlPayload := map[string]string{
		"webhook_url": mockServer.URL,
	}
	updateUrlBody, err := json.Marshal(updateUrlPayload)
	require.NoError(t, err)
	req, err = http.NewRequest("PUT", pPartApiURL+"/api/v1/org/"+orgName+"/partners/"+partnerID, bytes.NewBuffer(updateUrlBody))
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	req, err = http.NewRequest("POST", pPartApiURL+"/api/v1/org/"+orgName+"/partners/"+partnerID+"/verify", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var verifyResp map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&verifyResp)
	require.NoError(t, err)
	assert.Equal(t, "certified", verifyResp["status"])

	// 5. Delete partner
	req, err = http.NewRequest("DELETE", pPartApiURL+"/api/v1/org/"+orgName+"/partners/"+partnerID, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	// 6. List partners again to ensure deletion
	req, err = http.NewRequest("GET", pPartApiURL+"/api/v1/org/"+orgName+"/partners", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var listResp2 []map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&listResp2)
	require.NoError(t, err)
	assert.Len(t, listResp2, 0)
}
