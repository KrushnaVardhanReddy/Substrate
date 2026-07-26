package server

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
)

//go:embed all:static
var staticFiles embed.FS

// ServeDashboard serves the SvelteKit static build.
func ServeDashboard(r chi.Router) {
	staticSubFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		panic(err) // Should never happen in build
	}

	staticFS := http.FileServer(http.FS(staticSubFS))

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// If the request is for the API, it should not be handled by this Catch-All.
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		// Clean the path to check if it exists in the embedded FS
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Try to open the file in the embedded FS
		file, err := staticSubFS.Open(path)
		if err == nil {
			file.Close()
			staticFS.ServeHTTP(w, r)
			return
		}

		// File does not exist, let's serve index.html for SPA routing
		// We intercept and serve the embedded index.html
		indexFile, err := staticSubFS.Open("index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer indexFile.Close()

		// If we let FileServer handle this with r.URL.Path = "/index.html",
		// it might 301 redirect if the original request was to a directory path.
		// Instead, we can just read and serve the contents.
		stat, err := indexFile.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}

		// Serve the file content
		http.ServeContent(w, r, "index.html", stat.ModTime(), indexFile.(io.ReadSeeker))
	})
}
