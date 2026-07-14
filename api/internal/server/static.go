package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

//go:embed dashboard_build/*
var dashboardBuild embed.FS

// ServeDashboard embeds the SvelteKit static build and serves it.
func ServeDashboard(r chi.Router) {
	// Create a sub filesystem that points directly inside the build directory
	fsys, err := fs.Sub(dashboardBuild, "dashboard_build")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(fsys))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// If the request is for the API, it should not be handled by this Catch-All.
		// However, chi routes are evaluated in order of registration (with some tree-routing magic).
		// By registering this last or using a catch-all, chi handles it.
		// Just to be safe, if we get here and it's an /api/ route, return 404
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Try to serve the exact file
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		_, err := fsys.Open(path)
		if err != nil {
			// File does not exist, let's serve index.html for SPA routing
			r.URL.Path = "/"
		}

		fileServer.ServeHTTP(w, r)
	})
}
