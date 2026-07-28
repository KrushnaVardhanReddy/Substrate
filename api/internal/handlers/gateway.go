package handlers

import (
	"context"
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

		w.WriteHeader(http.StatusAccepted)

		go func() {
			err := gateway.RunGatewaySync(context.Background(), store, ghClient, org, repo)
			if err != nil {
				log.Printf("Gateway sync failed for %s/%s: %v", org, repo, err)
			}
		}()
	}
}
