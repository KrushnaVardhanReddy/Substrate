package server

import (
	"fmt"
	"net/http"
	"strings"

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

			orgsClaim, ok := claims["orgs"].([]interface{})
			if !ok {
				http.Error(w, "missing or invalid orgs claim", http.StatusForbidden)
				return
			}

			orgsMap := make(map[string]bool)
			for _, org := range orgsClaim {
				if orgStr, ok := org.(string); ok {
					orgsMap[orgStr] = true
				}
			}

			// Extract {org} from path (naive extraction for now, usually mux provides this if standard 1.22 routing is used)
			// Our paths are like /api/v1/repos/{org}, /api/v1/graph/{org}, /api/v1/schema/{org}/{repo}

			// We can use the Request's PathValue method which is available in Go 1.22+ ServeMux
			orgPath := r.PathValue("org")
			if orgPath != "" && !orgsMap[orgPath] {
				http.Error(w, "forbidden: unallowed org access", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
