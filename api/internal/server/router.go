package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/handlers"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/registry"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/services"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/webhook"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/workers"
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
func NewRouter(store db.Store, riverClient workers.JobEnqueuer, authConfig handlers.AuthConfig, registryApiToken, jwtSecret string) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS", "PUT", "DELETE"},
		AllowedHeaders: []string{"*"},
	}))

	r.Get("/health", handlers.HealthHandler)
	// Unprotected Auth routes
	r.Get("/api/v1/auth/github/login", handlers.HandleGitHubLogin(authConfig))
	r.Get("/api/v1/auth/github/callback", handlers.HandleGitHubCallback(authConfig))

	// Stripe Webhook (unprotected, validates Stripe signature internally)
	r.Method("POST", "/api/v1/webhooks/stripe", http.HandlerFunc(handlers.StripeWebhookHandler(store)))

	// Protected routes (Service Token only)
	serviceTokenMW := ServiceTokenMiddleware(registryApiToken)
	limitsMW := TierLimitsMiddleware(store)

	r.Method("POST", "/api/v1/sync", serviceTokenMW(limitsMW(http.HandlerFunc(handlers.SyncHandler(store, riverClient)))))
	r.Method("POST", "/api/v1/webhook", serviceTokenMW(http.HandlerFunc(webhook.PushHandler(store, github.NewRESTClient(), riverClient))))
	r.Method("POST", "/api/v1/webhook/reaction", serviceTokenMW(http.HandlerFunc(webhook.ReactionHandler(store, github.NewRESTClient()))))
	r.Method("POST", "/api/v1/cross-repo-check", serviceTokenMW(limitsMW(http.HandlerFunc(handlers.CrossRepoCheckHandler(store, riverClient)))))
	r.Method("POST", "/api/v1/history", serviceTokenMW(http.HandlerFunc(handlers.HistoryHandler(store))))
	r.Method("GET", "/api/v1/changes", serviceTokenMW(http.HandlerFunc(handlers.ChangesHandler(store))))
	r.Method("GET", "/api/v1/registry/can-deploy", serviceTokenMW(http.HandlerFunc(handlers.CanDeployHandler(store))))
	r.Method("GET", "/api/v1/registry/can-rollback", serviceTokenMW(http.HandlerFunc(handlers.CanRollbackHandler(store))))
	r.Method("POST", "/api/v1/diff", serviceTokenMW(http.HandlerFunc(handlers.SaveDiffHandler(store, riverClient, github.NewRESTClient()))))

	// Protected routes (Service Token OR JWT)
	authMW := AuthMiddleware(registryApiToken, jwtSecret)
	authzMW := AuthzMiddleware(registryApiToken, jwtSecret)
	r.Method("POST", "/api/v1/org/{org}/webhooks", authzMW(http.HandlerFunc(handlers.RegisterWebhookHandler(store))))
	r.Method("GET", "/api/v1/graph/{org}", authMW(http.HandlerFunc(handlers.GraphHandler(store))))
	r.Method("GET", "/api/v1/export/docs/{org}", authMW(http.HandlerFunc(handlers.HandleExportDocs(store))))
	r.Method("GET", "/api/v1/impact/{org}/{repo}", authMW(http.HandlerFunc(handlers.ImpactHandler(store))))
	r.Method("GET", "/api/v1/repos/{org}", authMW(http.HandlerFunc(handlers.ReposHandler(store))))
	r.Method("GET", "/api/v1/schema/{owner}/{repo}", authMW(http.HandlerFunc(handlers.SchemaHandler(store))))
	r.Method("GET", "/api/v1/history/{org}/{repo}", authMW(http.HandlerFunc(handlers.HistoryGetHandler(store))))
	r.Method("POST", "/api/v1/telemetry/traces", serviceTokenMW(http.HandlerFunc(handlers.TelemetryHandler(store))))
	r.Method("POST", "/api/v1/telemetry/drift", serviceTokenMW(http.HandlerFunc(handlers.DriftTelemetryHandler(store))))
	r.Method("GET", "/api/v1/telemetry/roi/{org}", authzMW(http.HandlerFunc(handlers.ROIHandler(store))))

	// SSE endpoint for live graph updates
	r.Method("GET", "/api/v1/events", authMW(http.HandlerFunc(handlers.EventsHandler)))

	// Route uses GitHub OAuth token directly, not the internal JWT, so we skip authMW.
	// The endpoint validates the token by making a call to GitHub.
	r.Method("POST", "/api/v1/org/{org}/enforce", authzMW(http.HandlerFunc(handlers.EnforceGlobalHandler())))

	// FinOps routes
	r.Method("POST", "/api/v1/finops/predict", serviceTokenMW(http.HandlerFunc(handlers.HandlePredictCost)))

	// AI routes (public — no auth required, BYOK model)
	r.Post("/api/v1/ai/analyze", services.AIAnalyzeHandler())
	r.Post("/api/v1/ai/autofix", services.AIAutofixHandler())

	// Public registry routes
	r.Route("/api/v1/registry/public", func(r chi.Router) {
		r.Post("/{namespace}/{name}/{version}", registry.HandlePublishSchema(store))
		r.Get("/{namespace}/{name}/{version}", registry.HandleFetchSchema(store))
	})

	// Public routes
	r.Get("/api/v1/diff/{id}", handlers.GetDiffHandler(store))
	r.Get("/api/v1/changelog/{org}/{repo}", handlers.ChangelogHandler(store))

	// Webhook for Postman integrations
	ServeDashboard(r)
	return otelhttp.NewHandler(r, "substrate-api")
}
