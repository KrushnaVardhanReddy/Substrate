package handlers

import (
	"log"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/gateway"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/go-chi/chi/v5"
)

func GatewaySyncHandler(store db.Store, ghClient github.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := chi.URLParam(r, "org")
		repo := chi.URLParam(r, "repo")

		crdYaml, err := gateway.RunGatewaySync(r.Context(), store, ghClient, org, repo)
		if err != nil {
			log.Printf("Gateway sync failed for %s/%s: %v", org, repo, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/x-yaml")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(crdYaml))
	}
}
