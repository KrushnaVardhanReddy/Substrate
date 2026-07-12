package server

import (
	"net/http"

	"github.com/KrushnaVardhanReddy/Substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/Substrate/api/internal/handlers"
	"github.com/KrushnaVardhanReddy/Substrate/api/internal/webhook"
)

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func ServiceTokenMiddleware(registryApiToken string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || authHeader != "Bearer "+registryApiToken {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// NewRouter creates a new router with all the routes registered.
func NewRouter(store db.Store, authConfig handlers.AuthConfig, registryApiToken, jwtSecret string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.HealthHandler)
	// Unprotected Auth routes
	mux.HandleFunc("GET /api/v1/auth/github/login", handlers.HandleGitHubLogin(authConfig))
	mux.HandleFunc("GET /api/v1/auth/github/callback", handlers.HandleGitHubCallback(authConfig))

	// Protected routes (Service Token only)
	serviceTokenMW := ServiceTokenMiddleware(registryApiToken)
	limitsMW := TierLimitsMiddleware(store)

	mux.Handle("POST /api/v1/sync", serviceTokenMW(limitsMW(http.HandlerFunc(handlers.SyncHandler(store)))))
	mux.Handle("POST /api/v1/cross-repo-check", serviceTokenMW(limitsMW(http.HandlerFunc(handlers.CrossRepoCheckHandler(store)))))
	mux.Handle("POST /api/v1/history", serviceTokenMW(http.HandlerFunc(handlers.HistoryHandler(store))))

	// Protected routes (Service Token OR JWT)
	authMW := AuthMiddleware(registryApiToken, jwtSecret)
	mux.Handle("GET /api/v1/graph/{org}", authMW(http.HandlerFunc(handlers.GraphHandler(store))))
	mux.Handle("GET /api/v1/repos/{org}", authMW(http.HandlerFunc(handlers.ReposHandler(store))))
	mux.Handle("GET /api/v1/schema/{owner}/{repo}", authMW(http.HandlerFunc(handlers.SchemaHandler(store))))
	mux.Handle("GET /api/v1/history/{org}/{repo}", authMW(http.HandlerFunc(handlers.HistoryGetHandler(store))))

	// AI routes (public — no auth required, BYOK model)
	mux.HandleFunc("POST /api/v1/ai/analyze", handlers.AIAnalyzeHandler())
	mux.HandleFunc("POST /api/v1/ai/autofix", handlers.AIAutofixHandler())

	// Webhook for Postman integrations
	mux.HandleFunc("POST /api/v1/webhook", webhook.Handler())

	return corsMiddleware(mux)
}
