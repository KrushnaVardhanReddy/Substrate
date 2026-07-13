package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/discovery"
)

func TelemetryHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var spans []discovery.OTelSpan
		if err := json.NewDecoder(r.Body).Decode(&spans); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		// Asynchronous processing to prevent blocking the client
		// We copy the spans and use context.Background() since the request context will be cancelled when we return.
		go func(spansCopy []discovery.OTelSpan) {
			err := discovery.ProcessRuntimeSignals(context.Background(), store, spansCopy)
			if err != nil {
				log.Printf("failed to process telemetry: %v", err)
			}
		}(spans)

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
	}
}

func DriftTelemetryHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var anomaly db.DriftAnomaly
		if err := json.NewDecoder(r.Body).Decode(&anomaly); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if anomaly.OrgName == "" || anomaly.RepoName == "" || anomaly.Path == "" {
			http.Error(w, "missing required fields", http.StatusBadRequest)
			return
		}

		if err := store.RecordDriftAnomaly(r.Context(), anomaly); err != nil {
			log.Printf("failed to record drift anomaly: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
	}
}
