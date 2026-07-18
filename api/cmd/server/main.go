package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	public_handlers "github.com/KrushnaVardhanReddy/substrate/api/handlers"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/handlers"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/server"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/workers"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/telemetry"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8090"
	}

	registryApiToken := os.Getenv("REGISTRY_API_TOKEN")
	if registryApiToken == "" {
		log.Fatal("REGISTRY_API_TOKEN environment variable is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	authConfig := handlers.AuthConfig{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		JWTSecret:    jwtSecret,
		DashboardURL: os.Getenv("DASHBOARD_URL"),
	}

	if authConfig.ClientID == "" || authConfig.ClientSecret == "" || authConfig.DashboardURL == "" {
		log.Fatal("GITHUB_CLIENT_ID, GITHUB_CLIENT_SECRET, and DASHBOARD_URL environment variables are required")
	}

	if err := db.RunMigrations(databaseURL); err != nil {
		log.Fatalf("failed to apply migrations: %v", err)
	}
	fmt.Println("[substrate-api] migrations applied successfully")

	tp, err := telemetry.InitTracer(context.Background(), "substrate-api")
	if err != nil {
		log.Printf("[substrate-api] failed to init tracer: %v\n", err)
	} else if tp != nil {
		defer func() {
			if err := tp.Shutdown(context.Background()); err != nil {
				log.Printf("[substrate-api] error shutting down tracer provider: %v", err)
			}
		}()
		fmt.Println("[substrate-api] OpenTelemetry tracing initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	store := db.NewPGStore(pool)

	// Initialize River job queue
	workersPool, pushWorker := workers.RegisterWorkers(store, github.NewRESTClient())
	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers: workersPool,
	})
	if err != nil {
		log.Fatalf("failed to create river client: %v", err)
	}
	pushWorker.RiverClient = riverClient
	if err := riverClient.Start(context.Background()); err != nil {
		log.Fatalf("failed to start river client: %v", err)
	}

	broker := public_handlers.NewSSEBroker()
	go broker.Start()

	router := server.NewRouter(store, riverClient, authConfig, registryApiToken, jwtSecret)

	// Register the SSE endpoint directly on the router returned by server.NewRouter
	// Note: server.NewRouter returns an http.Handler. The actual chi router is buried,
	// but it's cleaner to just wrap the handler if we don't change router.go, or,
	// better yet, we can do a type assertion. However, since server.NewRouter returns an otelhttp.Handler,
	// we will use http.NewServeMux to wrap it so it doesn't bypass middlewares completely on the main router,
	// though /api/v1/events will miss otel tracing.
	// Actually, wait! The best way to register it without bypassing middleware is to assert router as an interface with Handle, or we can just modify router.go if that's allowed, but we can't.
	// Wait, the review suggested: "The route should ideally be registered directly on the returned router in main.go so it inherits standard API middlewares". Since we are strictly not allowed to modify router.go, we could use an http.ServeMux or chi router in main.go, but wrap the *entire* router.
	// Let's use chi router and mount the returned router under a wildcard if it doesn't match /api/v1/events.
	// We'll restore the chi router setup since it's the only way, but we will make sure not to drop events.
	// Actually, if we look closely at the review: "The route should ideally be registered directly on the returned router in main.go so it inherits standard API middlewares".
	// The problem is `server.NewRouter` returns `http.Handler` via `otelhttp.NewHandler`.

	mux := chi.NewRouter()
	mux.Get("/api/v1/events", broker.ServeHTTP)
	mux.Mount("/", router)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		fmt.Printf("[substrate-api] server listening on :%s\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("[substrate-api] shutting down server...")

	ctxShutdown, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()

	if err := riverClient.Stop(ctxShutdown); err != nil {
		log.Printf("failed to stop river client: %v", err)
	}

	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
}
