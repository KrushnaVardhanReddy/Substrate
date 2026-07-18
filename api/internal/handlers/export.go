package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"github.com/go-chi/chi/v5"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

// HandleExportDocs handles the /api/v1/export/docs/{org} endpoint.
func HandleExportDocs(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := chi.URLParam(r, "org")
		if org == "" {
			http.Error(w, "org missing", http.StatusBadRequest)
			return
		}

		edges, err := store.GetDependencyGraph(r.Context(), org)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		var sb strings.Builder
		sb.WriteString("# Substrate Architecture\n\n")
		sb.WriteString("```mermaid\n")
		sb.WriteString("graph TD\n")

		for _, edge := range edges {
			sb.WriteString(fmt.Sprintf("  \"%s\" --> \"%s\"\n", edge.ConsumerFullName, edge.ProviderFullName))
		}

		sb.WriteString("```\n")

		w.Header().Set("Content-Type", "text/markdown")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(sb.String()))
	}
}
