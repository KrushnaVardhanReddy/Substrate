package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

func TestAuthMiddleware(t *testing.T) {
	registryToken := "test-registry-token"
	jwtSecret := "test-jwt-secret"

	// Generate a valid JWT token
	validToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"orgs": []string{"allowed-org"},
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	validTokenString, _ := validToken.SignedString([]byte(jwtSecret))

	// Generate a valid JWT token but expired
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"orgs": []string{"allowed-org"},
		"exp":  time.Now().Add(-time.Hour).Unix(),
	})
	expiredTokenString, _ := expiredToken.SignedString([]byte(jwtSecret))

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
