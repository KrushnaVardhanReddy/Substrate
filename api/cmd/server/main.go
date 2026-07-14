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

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/handlers"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/server"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/workers"
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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	store := db.NewPGStore(pool)

	// Initialize River job queue
	workersPool := workers.RegisterWorkers(store, github.NewRESTClient())
	riverClient, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers: workersPool,
	})
	if err != nil {
		log.Fatalf("failed to create river client: %v", err)
	}
	if err := riverClient.Start(ctx); err != nil {
		log.Fatalf("failed to start river client: %v", err)
	}

	router := server.NewRouter(store, riverClient, authConfig, registryApiToken, jwtSecret)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
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
