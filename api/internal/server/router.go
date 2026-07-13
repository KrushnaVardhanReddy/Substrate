package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/handlers"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/webhook"
)

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
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},
		AllowedHeaders:   []string{"*"},
	}))

	r.Get("/health", handlers.HealthHandler)
	// Unprotected Auth routes
	r.Get("/api/v1/auth/github/login", handlers.HandleGitHubLogin(authConfig))
	r.Get("/api/v1/auth/github/callback", handlers.HandleGitHubCallback(authConfig))

	// Protected routes (Service Token only)
	serviceTokenMW := ServiceTokenMiddleware(registryApiToken)
	limitsMW := TierLimitsMiddleware(store)

	r.Method("POST", "/api/v1/sync", serviceTokenMW(limitsMW(http.HandlerFunc(handlers.SyncHandler(store)))))
	r.Method("POST", "/api/v1/webhook", serviceTokenMW(http.HandlerFunc(webhook.PushHandler(store))))
	r.Method("POST", "/api/v1/cross-repo-check", serviceTokenMW(limitsMW(http.HandlerFunc(handlers.CrossRepoCheckHandler(store)))))
	r.Method("POST", "/api/v1/history", serviceTokenMW(http.HandlerFunc(handlers.HistoryHandler(store))))
	r.Method("GET", "/api/v1/registry/can-deploy", serviceTokenMW(http.HandlerFunc(handlers.CanDeployHandler(store))))
	r.Method("POST", "/api/v1/diff", serviceTokenMW(http.HandlerFunc(handlers.SaveDiffHandler(store))))

	// Protected routes (Service Token OR JWT)
	authMW := AuthMiddleware(registryApiToken, jwtSecret)
	r.Method("POST", "/api/v1/org/{org}/webhooks", authMW(http.HandlerFunc(handlers.RegisterWebhookHandler(store))))
	r.Method("GET", "/api/v1/graph/{org}", authMW(http.HandlerFunc(handlers.GraphHandler(store))))
	r.Method("GET", "/api/v1/repos/{org}", authMW(http.HandlerFunc(handlers.ReposHandler(store))))
	r.Method("GET", "/api/v1/schema/{owner}/{repo}", authMW(http.HandlerFunc(handlers.SchemaHandler(store))))
	r.Method("GET", "/api/v1/history/{org}/{repo}", authMW(http.HandlerFunc(handlers.HistoryGetHandler(store))))
	r.Method("POST", "/api/v1/telemetry/traces", serviceTokenMW(http.HandlerFunc(handlers.TelemetryHandler(store))))

	// AI routes (public — no auth required, BYOK model)
	r.Post("/api/v1/ai/analyze", handlers.AIAnalyzeHandler())
	r.Post("/api/v1/ai/autofix", handlers.AIAutofixHandler())

	// Public routes
	r.Get("/api/v1/diff/{id}", handlers.GetDiffHandler(store))

	// Webhook for Postman integrations
	return r
}
