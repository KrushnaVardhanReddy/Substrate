package server

import (
	"net/http"

	"github.com/KrushnaVardhanReddy/Substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/Substrate/api/internal/handlers"
)

// NewRouter creates a new router with all the routes registered.
func NewRouter(store db.Store) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.HealthHandler)
	mux.HandleFunc("POST /api/v1/sync", handlers.SyncHandler(store))
	mux.HandleFunc("POST /api/v1/cross-repo-check", handlers.CrossRepoCheckHandler(store))
	mux.HandleFunc("GET /api/v1/graph/{org}", handlers.GraphHandler(store))
	mux.HandleFunc("GET /api/v1/repos/{org}", handlers.ReposHandler(store))

	return mux
}
