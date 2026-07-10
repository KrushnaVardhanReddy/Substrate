package server

import (
	"net/http"

	"github.com/KrushnaVardhanReddy/Substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/Substrate/api/internal/handlers"
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

// NewRouter creates a new router with all the routes registered.
func NewRouter(store db.Store) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handlers.HealthHandler)
	mux.HandleFunc("POST /api/v1/sync", handlers.SyncHandler(store))
	mux.HandleFunc("POST /api/v1/cross-repo-check", handlers.CrossRepoCheckHandler(store))
	mux.HandleFunc("GET /api/v1/graph/{org}", handlers.GraphHandler(store))
	mux.HandleFunc("GET /api/v1/repos/{org}", handlers.ReposHandler(store))
	mux.HandleFunc("GET /api/v1/schema/{owner}/{repo}", handlers.SchemaHandler(store))

	return corsMiddleware(mux)
}
