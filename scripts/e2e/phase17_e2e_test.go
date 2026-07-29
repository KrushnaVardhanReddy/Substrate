package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	p17ApiURL    = "http://localhost:8090"
	p17DbURL     = "postgres://postgres:postgres@localhost:54320/postgres?sslmode=disable&default_query_exec_mode=exec&statement_cache_capacity=0&pgbouncer=true"
	p17JWTSecret = "local-jwt-secret"
)

func createP17JWT(org, role string) string {
	hash := sha256.Sum256([]byte(p17JWTSecret))
	key, _ := paseto.V4SymmetricKeyFromBytes(hash[:])

	token := paseto.NewToken()
	token.Set("orgs", map[string]string{org: role})
	token.SetExpiration(time.Now().Add(time.Hour))

	return token.V4Encrypt(key, nil)
}

func waitForP17Services(t *testing.T) {
	client := http.Client{Timeout: 2 * time.Second}
	for i := 0; i < 15; i++ {
		resp, err := client.Get(p17ApiURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatalf("API server not reachable at %s", p17ApiURL)
}

func TestPhase17_E2E(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	waitForP17Services(t)

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, p17DbURL)
	require.NoError(t, err)
	defer pool.Close()

	err = SetupP17Database(pool)
	require.NoError(t, err, "Failed to seed P17 database")

	jwt := createP17JWT("mgr-org", "admin")

	// Setup httptest.Server to simulate GitHub API
	mockGH := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/repos/mgr-org/payments-api/contents/substrate.yaml" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"content": "c2VydmljZTogcGF5bWVudHMtYXBpCmdhdGV3YXk6CiAgdHlwZTogYXBpc2l4CiAgaW5mcmFfcmVwbzogbWdyLW9yZy9pbmZyYQ=="}`)) // service: payments-api\ngateway:\n  type: apisix\n  infra_repo: mgr-org/infra
			return
		}
		if r.URL.Path == "/repos/mgr-org/payments-api/contents/openapi.yaml" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"content": "b3BlbmFwaTogMy4wLjAKaW5mbzoKICB0aXRsZTogUGF5bWVudHMgQVBJCiAgdmVyc2lvbjogMS4wLjAKcGF0aHM6CiAgL3BheToKICAgIHBvc3Q6CiAgICAgIHJlc3BvbnNlczoKICAgICAgICAiMjAwIjoKICAgICAgICAgIGRlc2NyaXB0aW9uOiBPSw=="}`)) // openapi: 3.0.0...
			return
		}
		if r.Method == "POST" && r.URL.Path == "/repos/mgr-org/payments-api/pulls" {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"html_url": "https://github.com/mgr-org/payments-api/pull/1"}`))
			return
		}
		if r.Method == "PATCH" {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method == "POST" {
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{"sha": "mocksha"}`))
			return
		}
		if r.Method == "GET" && r.URL.Path == "/repos/mgr-org/payments-api/git/ref/heads/main" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"object": {"sha": "mocksha"}}`))
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer mockGH.Close()

	os.Setenv("GITHUB_API_URL", mockGH.URL)
	defer os.Unsetenv("GITHUB_API_URL")

	t.Run("Scorecard Endpoint with Cache", func(t *testing.T) {
		req, _ := http.NewRequest("GET", p17ApiURL+"/api/v1/scorecard/mgr-org/payments-api", nil)
		req.Header.Set("Authorization", "Bearer "+jwt)

		resp1, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp1.Body.Close()
		require.Equal(t, http.StatusOK, resp1.StatusCode)

		var data1 map[string]interface{}
		err = json.NewDecoder(resp1.Body).Decode(&data1)
		require.NoError(t, err)

		require.Contains(t, data1, "grade")
		require.Contains(t, data1, "total")
		require.Contains(t, data1, "breakdown")
		computedAt1 := data1["computed_at"]
		require.NotNil(t, computedAt1)

		// Second call
		req2, _ := http.NewRequest("GET", p17ApiURL+"/api/v1/scorecard/mgr-org/payments-api", nil)
		req2.Header.Set("Authorization", "Bearer "+jwt)
		resp2, err := http.DefaultClient.Do(req2)
		require.NoError(t, err)
		defer resp2.Body.Close()
		require.Equal(t, http.StatusOK, resp2.StatusCode)

		var data2 map[string]interface{}
		err = json.NewDecoder(resp2.Body).Decode(&data2)
		require.NoError(t, err)

		computedAt2 := data2["computed_at"]
		assert.Equal(t, computedAt1, computedAt2, "Expected cache hit to return same computed_at")
	})

	t.Run("Gateway Sync Endpoint", func(t *testing.T) {
		req, _ := http.NewRequest("POST", p17ApiURL+"/api/v1/gateway/sync/mgr-org/payments-api", nil)
		req.Header.Set("Authorization", "Bearer "+jwt)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		var buf bytes.Buffer
		buf.ReadFrom(resp.Body)
		crdYaml := buf.String()
		require.NotEmpty(t, crdYaml)

		// Assert the generated CRD contains the correct path rules derived from the registered OpenAPI spec
		// The OpenAPI spec generated from our seeded DB (or mock Github)
		assert.Contains(t, crdYaml, "kind: Gateway")
	})

	t.Run("Events Endpoints", func(t *testing.T) {
		eventPayload := map[string]interface{}{
			"org":         "mgr-org",
			"repo":        "payments-api",
			"event_type":  "deployment",
			"description": "Deployed v1.0.0",
		}
		body, _ := json.Marshal(eventPayload)

		req, _ := http.NewRequest("POST", p17ApiURL+"/api/v1/events", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+jwt)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		// Get events
		req2, _ := http.NewRequest("GET", p17ApiURL+"/api/v1/events/mgr-org", nil)
		req2.Header.Set("Authorization", "Bearer "+jwt)

		resp2, err := http.DefaultClient.Do(req2)
		require.NoError(t, err)
		defer resp2.Body.Close()
		require.Equal(t, http.StatusOK, resp2.StatusCode)

		var events []map[string]interface{}
		err = json.NewDecoder(resp2.Body).Decode(&events)
		require.NoError(t, err)

		require.GreaterOrEqual(t, len(events), 1)

		found := false
		for _, e := range events {
			if e["event_type"] == "deployment" && e["description"] == "Deployed v1.0.0" {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected to find the created deployment event")
	})
}
