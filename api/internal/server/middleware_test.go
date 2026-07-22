package server

import (
	"context"
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/go-chi/chi/v5"
)

func TestAuthMiddleware(t *testing.T) {
	registryToken := "test-registry-token"
	jwtSecret := "test-jwt-secret"

	hash := sha256.Sum256([]byte(jwtSecret))
	key, _ := paseto.V4SymmetricKeyFromBytes(hash[:])

	// Generate a valid PASETO token
	validToken := paseto.NewToken()
	validToken.SetExpiration(time.Now().Add(time.Hour))
	validToken.Set("orgs", map[string]string{"allowed-org": "member"})
	validTokenString := validToken.V4Encrypt(key, nil)

	// Generate an expired PASETO token
	expiredToken := paseto.NewToken()
	expiredToken.SetExpiration(time.Now().Add(-time.Hour))
	expiredToken.Set("orgs", map[string]string{"allowed-org": "member"})
	expiredTokenString := expiredToken.V4Encrypt(key, nil)

	handler := AuthMiddleware(registryToken, jwtSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name           string
		method         string
		path           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "missing authorization header",
			method:         "GET",
			path:           "/api/v1/repos/allowed-org",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid authorization format",
			method:         "GET",
			path:           "/api/v1/repos/allowed-org",
			authHeader:     "Basic dXNlcm5hbWU6cGFzc3dvcmQ=",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "service token success",
			method:         "GET",
			path:           "/api/v1/repos/allowed-org",
			authHeader:     "Bearer " + registryToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "service token success POST",
			method:         "POST",
			path:           "/api/v1/sync",
			authHeader:     "Bearer " + registryToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "jwt token success GET",
			method:         "GET",
			path:           "/api/v1/repos/allowed-org",
			authHeader:     "Bearer " + validTokenString,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "jwt token forbidden POST",
			method:         "POST",
			path:           "/api/v1/sync",
			authHeader:     "Bearer " + validTokenString,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "jwt token forbidden unallowed org GET",
			method:         "GET",
			path:           "/api/v1/repos/unallowed-org",
			authHeader:     "Bearer " + validTokenString,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "invalid jwt token signature",
			method:         "GET",
			path:           "/api/v1/repos/allowed-org",
			authHeader:     "Bearer " + validTokenString + "invalid",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "expired jwt token",
			method:         "GET",
			path:           "/api/v1/repos/allowed-org",
			authHeader:     "Bearer " + expiredTokenString,
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("org", extractOrgFromPath(tt.path))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}
		})
	}
}

// helper for tests to extract org from path like /api/v1/repos/{org}
func extractOrgFromPath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 5 {
		return parts[4]
	}
	return ""
}

func TestAuthzMiddleware(t *testing.T) {
	registryToken := "test-registry-token"
	jwtSecret := "test-jwt-secret"

	hash := sha256.Sum256([]byte(jwtSecret))
	key, _ := paseto.V4SymmetricKeyFromBytes(hash[:])

	// Generate a valid PASETO token with admin role
	adminToken := paseto.NewToken()
	adminToken.SetExpiration(time.Now().Add(time.Hour))
	adminToken.Set("orgs", map[string]string{"allowed-org": "admin"})
	adminTokenString := adminToken.V4Encrypt(key, nil)

	// Generate a valid PASETO token with member role
	memberToken := paseto.NewToken()
	memberToken.SetExpiration(time.Now().Add(time.Hour))
	memberToken.Set("orgs", map[string]string{"allowed-org": "member"})
	memberTokenString := memberToken.V4Encrypt(key, nil)

	handler := AuthzMiddleware(registryToken, jwtSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		name           string
		method         string
		path           string
		authHeader     string
		expectedStatus int
	}{
		{
			name:           "missing authorization header",
			method:         "POST",
			path:           "/api/v1/org/allowed-org/enforce",
			authHeader:     "",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "service token success",
			method:         "POST",
			path:           "/api/v1/org/allowed-org/enforce",
			authHeader:     "Bearer " + registryToken,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "jwt admin success",
			method:         "POST",
			path:           "/api/v1/org/allowed-org/enforce",
			authHeader:     "Bearer " + adminTokenString,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "jwt member forbidden",
			method:         "POST",
			path:           "/api/v1/org/allowed-org/enforce",
			authHeader:     "Bearer " + memberTokenString,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "jwt admin unallowed org",
			method:         "POST",
			path:           "/api/v1/org/unallowed-org/enforce",
			authHeader:     "Bearer " + adminTokenString,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("org", extractOrgFromPath(tt.path))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %v, got %v", tt.expectedStatus, rr.Code)
			}
		})
	}
}
