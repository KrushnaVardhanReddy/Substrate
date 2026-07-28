package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
)

type APIKeyHandler struct {
	Store ports.Store
}

func (h *APIKeyHandler) CreateAPIKey(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "org")
	orgID, err := h.Store.GetOrgIDByName(r.Context(), orgIDStr)
	if err != nil {
		http.Error(w, "organization not found", http.StatusNotFound)
		return
	}

	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}

	// Generate a secure 32-byte token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}
	rawToken := hex.EncodeToString(tokenBytes)

	// Hash the token
	hash, err := bcrypt.GenerateFromPassword([]byte(rawToken), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "failed to hash token", http.StatusInternalServerError)
		return
	}

	prefix := rawToken[:4] + "..."

	key, err := h.Store.CreateAPIKey(r.Context(), orgID, req.Name, prefix, string(hash))
	if err != nil {
		http.Error(w, "failed to create api key", http.StatusInternalServerError)
		return
	}

	// Return the raw token EXACTLY ONCE
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"key":       key,
		"raw_token": rawToken,
	})
}

func (h *APIKeyHandler) ListAPIKeys(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "org")
	orgID, err := h.Store.GetOrgIDByName(r.Context(), orgIDStr)
	if err != nil {
		http.Error(w, "organization not found", http.StatusNotFound)
		return
	}

	keys, err := h.Store.ListAPIKeys(r.Context(), orgID)
	if err != nil {
		http.Error(w, "failed to list api keys", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(keys)
}

func (h *APIKeyHandler) DeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "org")
	orgID, err := h.Store.GetOrgIDByName(r.Context(), orgIDStr)
	if err != nil {
		http.Error(w, "organization not found", http.StatusNotFound)
		return
	}

	keyIDStr := chi.URLParam(r, "id")
	keyID, err := uuid.Parse(keyIDStr)
	if err != nil {
		http.Error(w, "invalid key ID", http.StatusBadRequest)
		return
	}

	err = h.Store.DeleteAPIKey(r.Context(), keyID, orgID)
	if err != nil {
		http.Error(w, "failed to delete api key", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
