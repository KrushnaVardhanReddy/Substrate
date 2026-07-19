package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

// ServeDashboard serves the SvelteKit static build.
func ServeDashboard(r chi.Router) {
	staticFS := http.FileServer(http.Dir("./static"))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// If the request is for the API, it should not be handled by this Catch-All.
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Try to serve the exact file
		path := filepath.Join("./static", r.URL.Path)
		if _, err := os.Stat(path); os.IsNotExist(err) || path == "static" {
			// File does not exist (or root requested), let's serve index.html for SPA routing
			http.ServeFile(w, r, "./static/index.html")
			return
		}

		staticFS.ServeHTTP(w, r)
	})
}
