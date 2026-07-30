package server

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/handler"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/handlers"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/ai"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/marketplace"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
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
func NewRouter(store ports.Store, riverClient workers.JobEnqueuer, authConfig handlers.AuthConfig, registryApiToken, jwtSecret string) http.Handler {
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

	// Public Profile
	publicProfileHandler := handler.NewPublicProfileHandler(store.Pool())
	r.Get("/api/v1/public/profile/{org}", publicProfileHandler.GetPublicProfile)

	// Stripe Webhook (unprotected, validates Stripe signature internally)
	r.Method("POST", "/api/v1/webhooks/stripe", http.HandlerFunc(handlers.StripeWebhookHandler(store)))

	// Protected routes (Service Token only)
	serviceTokenMW := ServiceTokenMiddleware(registryApiToken)
	limitsMW := TierLimitsMiddleware(store)

	r.Method("POST", "/api/v1/sync", serviceTokenMW(limitsMW(http.HandlerFunc(handlers.SyncHandler(store, riverClient)))))
	r.Method("POST", "/api/v1/gateway/sync/{org}/{repo}", http.HandlerFunc(handlers.GatewaySyncHandler(store, github.NewRESTClient())))
	r.Method("POST", "/api/v1/webhook", http.HandlerFunc(webhook.PushHandler(store, github.NewRESTClient(), riverClient)))
	r.Method("POST", "/api/v1/webhook/reaction", serviceTokenMW(http.HandlerFunc(webhook.ReactionHandler(store, github.NewRESTClient()))))
	r.Method("POST", "/api/v1/cross-repo-check", serviceTokenMW(limitsMW(http.HandlerFunc(handlers.CrossRepoCheckHandler(store, riverClient)))))
	r.Method("POST", "/api/v1/history", serviceTokenMW(http.HandlerFunc(handlers.HistoryHandler(store))))
	r.Method("GET", "/api/v1/changes", serviceTokenMW(http.HandlerFunc(handlers.ChangesHandler(store))))
	r.Method("GET", "/api/v1/registry/can-deploy", serviceTokenMW(http.HandlerFunc(handlers.CanDeployHandler(store))))
	r.Method("GET", "/api/v1/registry/can-rollback", serviceTokenMW(http.HandlerFunc(handlers.CanRollbackHandler(store))))
	r.Method("POST", "/api/v1/diff", serviceTokenMW(http.HandlerFunc(handlers.SaveDiffHandler(store, riverClient, github.NewRESTClient()))))
	r.Method("POST", "/api/v1/postmortem", serviceTokenMW(http.HandlerFunc(handlers.PostmortemHandler(store))))
	r.Method("POST", "/api/v1/schema/smell", serviceTokenMW(http.HandlerFunc(handlers.SchemaSmellHandler())))
	r.Method("POST", "/api/v1/plugins/publish", serviceTokenMW(http.HandlerFunc(marketplace.PublishHandler(store))))

	// Risk score endpoint
	r.Method("GET", "/api/v1/risk/{org}/{repo}/{pr}", http.HandlerFunc(handlers.RiskScoreHandler()))

	r.Method("GET", "/api/v1/scorecard/{org}/{repo}", http.HandlerFunc(handlers.GetScorecardHandler(store)))

	// Guides docs API
	r.Method("GET", "/api/v1/docs/{org}/{repo}", http.HandlerFunc(handlers.ListGuidesHandler(store)))
	r.Method("GET", "/api/v1/docs/{org}/{repo}/{slug}", http.HandlerFunc(handlers.GetGuideHandler(store)))
	r.Method("GET", "/api/v1/docs/{org}/{repo}/{slug}/*", http.HandlerFunc(handlers.GetGuideHandler(store)))

	// Protected routes (Service Token OR JWT)
	authMW := AuthMiddleware(registryApiToken, jwtSecret)
	authzMW := AuthzMiddleware(registryApiToken, jwtSecret)

	// API Keys
	apiKeyHandler := &handlers.APIKeyHandler{Store: store}
	r.Method("POST", "/api/v1/org/{org}/apikeys", authzMW(http.HandlerFunc(apiKeyHandler.CreateAPIKey)))
	r.Method("GET", "/api/v1/org/{org}/apikeys", authzMW(http.HandlerFunc(apiKeyHandler.ListAPIKeys)))
	r.Method("DELETE", "/api/v1/org/{org}/apikeys/{id}", authzMW(http.HandlerFunc(apiKeyHandler.DeleteAPIKey)))

	// ArgoCD/Flux webhook
	r.Method("POST", "/api/v1/argo/drift-webhook", serviceTokenMW(http.HandlerFunc(handlers.ArgoDriftWebhookHandler(store))))

	// MCP Profiles and HITL Queue

	// MCP Audit Logs
	mcpAuditHandler := &handlers.MCPAuditHandler{Store: store}
	r.Method("GET", "/api/v1/ai/audit", authMW(http.HandlerFunc(mcpAuditHandler.ListMCPAuditLogs)))

	mcpProfileHandler := &handlers.MCPProfileHandler{Store: store}
	r.Method("POST", "/api/v1/mcp/profiles", authMW(http.HandlerFunc(mcpProfileHandler.CreateProfile)))
	r.Method("GET", "/api/v1/mcp/profiles/{org}", authzMW(http.HandlerFunc(mcpProfileHandler.ListProfiles)))
	r.Method("DELETE", "/api/v1/mcp/profiles/{id}", authMW(http.HandlerFunc(mcpProfileHandler.DeleteProfile)))
	r.Method("GET", "/api/v1/mcp/hitl-queue/{org}", authzMW(http.HandlerFunc(mcpProfileHandler.ListHITLQueue)))
	r.Method("POST", "/api/v1/mcp/hitl-queue/{id}/resolve", authMW(http.HandlerFunc(mcpProfileHandler.ResolveHITLItem)))

	sandboxHandler := &handlers.SandboxHandler{Store: store, JWTSecret: jwtSecret}
	r.Method("POST", "/api/v1/sandbox/token", authzMW(http.HandlerFunc(sandboxHandler.GenerateTokenHandler)))
	r.Method("POST", "/api/v1/sandbox/request", http.HandlerFunc(sandboxHandler.ProxyRequestHandler))

	jwtValidMW := JWTValidMiddleware(jwtSecret)

	partnersHandler := handler.NewPartnersHandler(store)
	r.Method("GET", "/api/v1/org/{org}/partners", authzMW(http.HandlerFunc(partnersHandler.ListPartners)))
	r.Method("POST", "/api/v1/org/{org}/partners", authzMW(http.HandlerFunc(partnersHandler.CreatePartner)))
	r.Method("PUT", "/api/v1/org/{org}/partners/{id}", authzMW(http.HandlerFunc(partnersHandler.UpdatePartner)))
	r.Method("DELETE", "/api/v1/org/{org}/partners/{id}", authzMW(http.HandlerFunc(partnersHandler.DeletePartner)))
	r.Method("POST", "/api/v1/org/{org}/partners/{id}/verify", authzMW(http.HandlerFunc(partnersHandler.VerifyPartner)))

	r.Method("POST", "/api/v1/org/{org}/webhooks", authzMW(http.HandlerFunc(handlers.RegisterWebhookHandler(store))))
	r.Method("GET", "/api/v1/graph/{org}", authMW(http.HandlerFunc(handlers.GraphHandler(store))))
	r.Method("GET", "/api/v1/export/docs/{org}", authMW(http.HandlerFunc(handlers.HandleExportDocs(store))))
	r.Method("GET", "/api/v1/impact/{org}/{repo}", authMW(http.HandlerFunc(handlers.ImpactHandler(store))))
	r.Method("GET", "/api/v1/repos/{org}", authMW(http.HandlerFunc(handlers.ReposHandler(store))))
	r.Method("GET", "/api/v1/schema/{owner}/{repo}", authMW(http.HandlerFunc(handlers.SchemaHandler(store))))
	r.Method("GET", "/api/v1/spec/{org}/{repo}", authMW(http.HandlerFunc(handlers.SpecHandler(store))))
	r.Method("GET", "/api/v1/history/{org}/{repo}", authMW(http.HandlerFunc(handlers.HistoryGetHandler(store))))
	r.Method("POST", "/api/v1/telemetry/track", http.HandlerFunc(handlers.TrackTelemetryHandler()))
	r.Method("POST", "/api/v1/telemetry/traces", serviceTokenMW(http.HandlerFunc(handlers.TelemetryHandler(store))))
	r.Method("POST", "/api/v1/telemetry/drift", serviceTokenMW(http.HandlerFunc(handlers.DriftTelemetryHandler(store))))
	r.Method("GET", "/api/v1/telemetry/roi/{org}", authzMW(http.HandlerFunc(handlers.ROIHandler(store))))

	// SSE endpoint for live graph updates
	r.Method("GET", "/api/v1/events", authMW(http.HandlerFunc(handlers.EventsHandler)))
	r.Method("POST", "/api/v1/events", authMW(http.HandlerFunc(handlers.CreateEventHandler(store))))
	r.Method("GET", "/api/v1/events/{org}", authMW(http.HandlerFunc(handlers.GetEventsHandler(store))))

	// OTel Metrics and Zombies (Phase 10)
	r.Method("POST", "/api/v1/otel/webhook", serviceTokenMW(http.HandlerFunc(handlers.OTelMetricsHandler(store))))
	r.Method("GET", "/api/v1/org/{org}/zombies", authzMW(http.HandlerFunc(handlers.GetZombiesHandler(store))))
	r.Method("POST", "/api/v1/org/{org}/zombies/pr", authzMW(http.HandlerFunc(handlers.CreateZombiePRHandler(github.NewRESTClient()))))

	// Governance Rules Generate CEL endpoint
	r.Method("POST", "/api/governance/generate-cel", jwtValidMW(handlers.GenerateCELHandler()))

	// Route uses GitHub OAuth token directly, not the internal JWT, so we skip authMW.
	// The endpoint validates the token by making a call to GitHub.
	r.Method("POST", "/api/v1/org/{org}/enforce", authzMW(http.HandlerFunc(handlers.EnforceGlobalHandler())))

	// Governance Rules API
	r.Method("POST", "/api/v1/enterprise/webhook", serviceTokenMW(http.HandlerFunc(webhook.EnterpriseWebhookPingHandler())))
	r.Method("POST", "/api/v1/enterprise/rules/validate", serviceTokenMW(http.HandlerFunc(handlers.ValidateCELRuleHandler())))
	r.Method("GET", "/api/v1/enterprise/drift/{org}/{repo}", serviceTokenMW(http.HandlerFunc(handlers.GetDriftReportHandler(store))))
	r.Method("POST", "/api/v1/enterprise/webhook", serviceTokenMW(http.HandlerFunc(webhook.EnterpriseWebhookPingHandler())))
	r.Method("POST", "/api/v1/enterprise/rules/validate", serviceTokenMW(http.HandlerFunc(handlers.ValidateCELRuleHandler())))
	r.Method("GET", "/api/v1/enterprise/drift/{org}/{repo}", serviceTokenMW(http.HandlerFunc(handlers.GetDriftReportHandler(store))))
	governanceHandler := handlers.NewGovernanceRulesHandler(store)
	r.Method("GET", "/api/v1/org/{org}/rules", authzMW(http.HandlerFunc(governanceHandler.ListRules)))
	r.Method("POST", "/api/v1/org/{org}/rules", authzMW(http.HandlerFunc(governanceHandler.CreateRule)))
	r.Method("DELETE", "/api/v1/org/{org}/rules/{ruleID}", authzMW(http.HandlerFunc(governanceHandler.DeleteRule)))
	r.Method("POST", "/api/v1/org/{org}/governance/rules", authzMW(http.HandlerFunc(governanceHandler.CreateRule)))

	// Insurance routes
	r.Method("GET", "/api/v1/org/{org}/insurance/policy", authzMW(http.HandlerFunc(handlers.InsuranceGetPolicyHandler(store))))
	r.Method("GET", "/api/v1/org/{org}/insurance/claims", authzMW(http.HandlerFunc(handlers.InsuranceGetClaimsHandler(store))))
	r.Method("POST", "/api/v1/org/{org}/insurance/claims", authzMW(http.HandlerFunc(handlers.InsuranceFileClaimHandler(store, github.NewRESTClient()))))

	// FinOps routes
	r.Method("POST", "/api/v1/finops/predict", serviceTokenMW(http.HandlerFunc(handlers.HandlePredictCost)))

	// AI routes (public — no auth required, BYOK model)
	r.Post("/api/v1/ai/analyze", services.AIAnalyzeHandler())
	r.Post("/api/v1/ai/autofix", services.AIAutofixHandler())
	r.Post("/api/v1/ai/impact", services.AIImpactHandler())

	// Phase 5 Discovery Routes
	r.Method("POST", "/api/v1/discovery/scan/{org}/{repo}", serviceTokenMW(http.HandlerFunc(handlers.TriggerScanHandler(riverClient))))
	r.Method("GET", "/api/v1/discovery/results/{org}/{repo}", serviceTokenMW(http.HandlerFunc(handlers.GetScanResultsHandler())))

	// Public registry routes
	r.Route("/api/v1/registry/public", func(r chi.Router) {
		r.Post("/{namespace}/{name}/{version}", registry.HandlePublishSchema(store))
		r.Get("/{namespace}/{name}/{version}", registry.HandleFetchSchema(store))
	})

	// Marketplace routes
	r.Method("POST", "/api/marketplace/publish", http.HandlerFunc(marketplace.PublishHandler(store)))
	r.Method("GET", "/api/marketplace/plugins", http.HandlerFunc(marketplace.ListPluginsHandler(store)))

	// Public routes
	r.Get("/api/v1/preview/{token}", handlers.GetPreviewHandler(store))
	r.Get("/api/v1/diff/{id}", handlers.GetDiffHandler(store))
	r.Get("/api/v1/changelog/{org}/{repo}", handlers.ChangelogHandler(store))
	r.Get("/api/badges/{org}/{repo}", handlers.BadgesHandler(store))

	// Integrations
	r.Method("POST", "/api/v1/integrations/pagerduty/webhook", http.HandlerFunc(handlers.PagerDutyWebhookHandler(store)))

	// Webhook for Postman integrations
	ServeDashboard(r)
	r.Method("GET", "/api/v1/qa/postman/{org}/{repo}", http.HandlerFunc(handlers.QAPostmanHandler(store)))
	r.Method("POST", "/api/v1/qa/shadow/replay", http.HandlerFunc(handlers.QAShadowReplayHandler(store)))
	r.Method("GET", "/api/v1/qa/coverage/{org}/{repo}", http.HandlerFunc(handlers.QACoverageHandler(store)))
	fuzzerHandler := &handlers.FuzzerHandler{Store: store}
	r.Method("GET", "/api/v1/fuzzer/gaps", http.HandlerFunc(fuzzerHandler.GetSchemaValidationGaps))

	// Phase 18 schema pruning
	aiClient := ai.NewClient()
	r.Method("POST", "/api/v1/schema/prune", http.HandlerFunc(handlers.SchemaPruneHandler(store, aiClient)))

	return otelhttp.NewHandler(r, "substrate-api")
}
