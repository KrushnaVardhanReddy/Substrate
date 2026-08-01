package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/crypto"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type AirbyteIngestRequest struct {
	Org     string            `json:"org"`
	Source  string            `json:"source"`
	Stream  string            `json:"stream"`
	Records []json.RawMessage `json:"records"`
}

type AirbyteIngestResponse struct {
	Ingested int `json:"ingested"`
}

type AirbyteSourceRequest struct {
	Name      string          `json:"name"`
	Connector string          `json:"connector"`
	Config    json.RawMessage `json:"config"`
}

type AirbyteHandler struct {
	store ports.AirbyteStore
}

func NewAirbyteHandler(store ports.AirbyteStore) *AirbyteHandler {
	return &AirbyteHandler{store: store}
}

func (h *AirbyteHandler) Ingest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req AirbyteIngestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Org == "" || req.Source == "" || req.Stream == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	source, err := h.store.GetAirbyteSource(ctx, sqlcgen.GetAirbyteSourceParams{
		Org:       req.Org,
		Connector: req.Source,
	})

	if err != nil {
		_, keyARN, err := h.store.GetOrgKMSConfig(ctx, req.Org)
		if err != nil || keyARN == "" {
			http.Error(w, "kms config not found for org", http.StatusBadRequest)
			return
		}
		kmsClient, err := crypto.NewKMSClient(keyARN)
		if err != nil {
			http.Error(w, "failed to initialize kms client", http.StatusInternalServerError)
			return
		}
		encConfig, _, err := kmsClient.Encrypt([]byte("{}"))
		if err != nil {
			http.Error(w, "failed to encrypt empty config", http.StatusInternalServerError)
			return
		}

		source, err = h.store.UpsertAirbyteSource(ctx, sqlcgen.UpsertAirbyteSourceParams{
			Org:       req.Org,
			Name:      req.Source,
			Connector: req.Source,
			ConfigEnc: encConfig,
		})
		if err != nil {
			http.Error(w, "failed to create source", http.StatusInternalServerError)
			return
		}
	}

	recordsBytes := make([][]byte, len(req.Records))
	for i, v := range req.Records {
		recordsBytes[i] = []byte(v)
	}

	ingested, err := h.store.BulkInsertAirbyteRecords(ctx, req.Org, source.ID.Bytes, req.Stream, recordsBytes)
	if err != nil {
		http.Error(w, "failed to bulk insert records", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(AirbyteIngestResponse{Ingested: ingested})
}

func (h *AirbyteHandler) CreateSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgName := chi.URLParam(r, "org")

	var req AirbyteSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	_, keyARN, err := h.store.GetOrgKMSConfig(ctx, orgName)
	if err != nil || keyARN == "" {
		http.Error(w, "kms config not found for org", http.StatusBadRequest)
		return
	}

	kmsClient, err := crypto.NewKMSClient(keyARN)
	if err != nil {
		http.Error(w, "failed to initialize kms client", http.StatusInternalServerError)
		return
	}

	encConfig, _, err := kmsClient.Encrypt(req.Config)
	if err != nil {
		http.Error(w, "failed to encrypt config", http.StatusInternalServerError)
		return
	}

	source, err := h.store.UpsertAirbyteSource(ctx, sqlcgen.UpsertAirbyteSourceParams{
		Org:       orgName,
		Name:      req.Name,
		Connector: req.Connector,
		ConfigEnc: encConfig,
	})
	if err != nil {
		http.Error(w, "failed to create source", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(source)
}

func (h *AirbyteHandler) ListSources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgName := chi.URLParam(r, "org")

	sources, err := h.store.ListAirbyteSources(ctx, orgName)
	if err != nil {
		http.Error(w, "failed to list sources", http.StatusInternalServerError)
		return
	}

	if sources == nil {
		sources = []sqlcgen.AirbyteSource{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sources)
}

func (h *AirbyteHandler) DeleteSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgName := chi.URLParam(r, "org")
	idParam := chi.URLParam(r, "id")

	id, err := uuid.Parse(idParam)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	err = h.store.DeleteAirbyteSource(ctx, sqlcgen.DeleteAirbyteSourceParams{
		Org: orgName,
		ID: pgtype.UUID{Bytes: id, Valid: true},
	})
	if err != nil {
		http.Error(w, "failed to delete source", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
