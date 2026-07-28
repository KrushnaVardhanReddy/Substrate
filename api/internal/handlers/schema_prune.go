package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/ai"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/schema"
	"github.com/getkin/kin-openapi/openapi3"
)

type PruneRequest struct {
	Org    string `json:"org"`
	Repo   string `json:"repo"`
	Intent string `json:"intent"`
}

func SchemaPruneHandler(store db.Store, aiClient ai.AIClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req PruneRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.Org == "" || req.Repo == "" || req.Intent == "" {
			http.Error(w, "org, repo, and intent are required", http.StatusBadRequest)
			return
		}

		maxEndpointsStr := r.URL.Query().Get("max_endpoints")
		maxEndpoints := 5
		if maxEndpointsStr != "" {
			parsed, err := strconv.Atoi(maxEndpointsStr)
			if err != nil {
				http.Error(w, "invalid max_endpoints", http.StatusBadRequest)
				return
			}
			maxEndpoints = parsed
		}

		if maxEndpoints < 1 || maxEndpoints > 20 {
			http.Error(w, "max_endpoints must be between 1 and 20", http.StatusBadRequest)
			return
		}

		hash := sha256.Sum256([]byte(req.Org + "/" + req.Repo + "/" + req.Intent))
		intentHash := hex.EncodeToString(hash[:])

		cached, err := store.GetPrunedSchemaCache(r.Context(), sqlcgen.GetPrunedSchemaCacheParams{
			Org:        req.Org,
			Repo:       req.Repo,
			IntentHash: intentHash,
		})

		if err == nil && len(cached.PrunedSchema) > 0 {
			w.Header().Set("Content-Type", "application/json")
			w.Write(cached.PrunedSchema)
			return
		}

		rawSchema, err := store.GetSchema(r.Context(), req.Org, req.Repo)
		if err != nil {
			if err == db.ErrNotFound {
				http.Error(w, "schema not found", http.StatusNotFound)
				return
			}
			http.Error(w, "failed to load schema", http.StatusInternalServerError)
			return
		}

		loader := openapi3.NewLoader()
		spec, err := loader.LoadFromData([]byte(rawSchema))
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to parse schema: %v", err), http.StatusInternalServerError)
			return
		}

		prunedSpec, err := schema.PruneSchema(r.Context(), spec, req.Intent, maxEndpoints, aiClient)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to prune schema: %v", err), http.StatusInternalServerError)
			return
		}

		prunedBytes, err := prunedSpec.MarshalJSON()
		if err != nil {
			http.Error(w, "failed to marshal pruned schema", http.StatusInternalServerError)
			return
		}

		err = store.UpsertPrunedSchemaCache(r.Context(), sqlcgen.UpsertPrunedSchemaCacheParams{
			Org:          req.Org,
			Repo:         req.Repo,
			IntentHash:   intentHash,
			PrunedSchema: prunedBytes,
		})
		if err != nil {
			fmt.Printf("failed to save pruned schema cache: %v\n", err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(prunedBytes)
	}
}
