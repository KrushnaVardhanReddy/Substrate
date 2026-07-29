package handlers_test

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
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/handlers"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/testutils"
)

func TestSandboxGenerateToken(t *testing.T) {
	pool, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	store := db.NewPGStore(pool)

	h := &handlers.SandboxHandler{
		Store:     store,
		JWTSecret: "test-secret",
	}

	reqBody, _ := json.Marshal(map[string]string{
		"org":  "mcp-org",
		"repo": "auth-service",
	})
	req := httptest.NewRequest("POST", "/api/v1/sandbox/token", bytes.NewReader(reqBody))
	rr := httptest.NewRecorder()
	h.GenerateTokenHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}

	var resp struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp.Token == "" {
		t.Fatal("expected token to not be empty")
	}
}

func TestSandboxSSRFPrevention(t *testing.T) {
	pool, cleanup := testutils.SetupTestDB(t)
	defer cleanup()
	store := db.NewPGStore(pool)

	orgID, _ := store.UpsertOrg(context.Background(), 123, "mcp-org")
	store.UpsertRepo(context.Background(), orgID, 456, "auth-service", "mcp-org/auth-service", nil)

	pool.Exec(context.Background(), "UPDATE repositories SET base_url = 'http://169.254.169.254' WHERE name = 'auth-service'")

	h := &handlers.SandboxHandler{
		Store:     store,
		JWTSecret: "test-secret",
	}

	hash := sha256.Sum256([]byte("test-secret"))
	key, _ := paseto.V4SymmetricKeyFromBytes(hash[:])
	token := paseto.NewToken()
	token.SetExpiration(time.Now().Add(15 * time.Minute))
	token.Set("org", "mcp-org")
	token.Set("repo", "auth-service")
	token.Set("sandbox", true)
	tokenString := token.V4Encrypt(key, nil)

	reqBody, _ := json.Marshal(map[string]string{
		"org":    "mcp-org",
		"repo":   "auth-service",
		"method": "GET",
		"path":   "/health",
	})
	req := httptest.NewRequest("POST", "/api/v1/sandbox/request", bytes.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rr := httptest.NewRecorder()

	h.ProxyRequestHandler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 Forbidden for private IP SSRF attempt, got %d: %s", rr.Code, rr.Body.String())
	}
}
