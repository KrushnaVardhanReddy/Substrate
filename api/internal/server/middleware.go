// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package server

import (
	"crypto/sha256"
	"net/http"
	"strings"

	"aidanwoods.dev/go-paseto"
	"github.com/go-chi/chi/v5"
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

			// Parse as PASETO v4 local
			hash := sha256.Sum256([]byte(jwtSecret))
			key, err := paseto.V4SymmetricKeyFromBytes(hash[:])
			if err != nil {
				http.Error(w, "internal server error: invalid key", http.StatusInternalServerError)
				return
			}

			parser := paseto.NewParser()
			parsedToken, err := parser.ParseV4Local(key, token, nil)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			var orgsClaim map[string]interface{}
			err = parsedToken.Get("orgs", &orgsClaim)
			if err != nil {
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

			// Parse as PASETO v4 local
			hash := sha256.Sum256([]byte(jwtSecret))
			key, err := paseto.V4SymmetricKeyFromBytes(hash[:])
			if err != nil {
				http.Error(w, "internal server error: invalid key", http.StatusInternalServerError)
				return
			}

			parser := paseto.NewParser()
			parsedToken, err := parser.ParseV4Local(key, token, nil)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			var orgsClaim map[string]interface{}
			err = parsedToken.Get("orgs", &orgsClaim)
			if err != nil {
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

func JWTValidMiddleware(jwtSecret string) func(http.Handler) http.Handler {
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

			hash := sha256.Sum256([]byte(jwtSecret))
			key, err := paseto.V4SymmetricKeyFromBytes(hash[:])
			if err != nil {
				http.Error(w, "internal server error: invalid key", http.StatusInternalServerError)
				return
			}

			parser := paseto.NewParser()
			_, err = parser.ParseV4Local(key, token, nil)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
