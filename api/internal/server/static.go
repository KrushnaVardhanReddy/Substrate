package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

// ServeDashboard serves the static SvelteKit build output.
func ServeDashboard(r chi.Router) {
	staticFS := http.FileServer(http.Dir("./static"))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		fullPath := filepath.Join("./static", path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			// SPA routing fallback
			http.ServeFile(w, r, "./static/index.html")
			return
		}

		staticFS.ServeHTTP(w, r)
	})
}
