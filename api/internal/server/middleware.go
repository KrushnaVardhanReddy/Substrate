package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(registryApiToken, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := parts[1]

			// If it exactly matches registryApiToken, allow full access
			if registryApiToken != "" && token == registryApiToken {
				next.ServeHTTP(w, r)
				return
			}

			// Reject non-GET requests if not using service token
			if r.Method != http.MethodGet {
				http.Error(w, "forbidden: write operations require service token", http.StatusForbidden)
				return
			}

			// Parse as JWT
			parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !parsedToken.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := parsedToken.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid claims", http.StatusUnauthorized)
				return
			}

			orgsClaim, ok := claims["orgs"].(map[string]interface{})
			if !ok {
				http.Error(w, "missing or invalid orgs claim", http.StatusForbidden)
				return
			}

			// Extract {org} from path
			orgPath := chi.URLParam(r, "org")
			if orgPath == "" {
				orgPath = chi.URLParam(r, "owner")
			}

			if orgPath != "" {
				if _, allowed := orgsClaim[orgPath]; !allowed {
					http.Error(w, "forbidden: unallowed org access", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func AuthzMiddleware(registryApiToken, jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "missing authorization header", http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}

			token := parts[1]

			// If it exactly matches registryApiToken, allow full access
			if registryApiToken != "" && token == registryApiToken {
				next.ServeHTTP(w, r)
				return
			}

			// Parse as JWT
			parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
				}
				return []byte(jwtSecret), nil
			})

			if err != nil || !parsedToken.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := parsedToken.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid claims", http.StatusUnauthorized)
				return
			}

			orgsClaim, ok := claims["orgs"].(map[string]interface{})
			if !ok {
				http.Error(w, "missing or invalid orgs claim", http.StatusForbidden)
				return
			}

			orgPath := chi.URLParam(r, "org")
			if orgPath == "" {
				http.Error(w, "bad request: missing org parameter", http.StatusBadRequest)
				return
			}

			role, ok := orgsClaim[orgPath].(string)
			if !ok || role != "admin" {
				http.Error(w, "forbidden: admin role required for this action", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
