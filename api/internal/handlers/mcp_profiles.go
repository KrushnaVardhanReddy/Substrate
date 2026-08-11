// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
	"github.com/go-chi/chi/v5"
)

type MCPProfileHandler struct {
	Store ports.Store
}

func (h *MCPProfileHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Org          string   `json:"org"`
		Name         string   `json:"name"`
		AllowedTools []string `json:"allowed_tools"`
		HITLEnabled  bool     `json:"hitl_enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Org == "" || req.Name == "" {
		http.Error(w, "org and name are required", http.StatusBadRequest)
		return
	}

	profile, err := h.Store.CreateAgentProfile(r.Context(), db.AgentProfile{
		Org:          req.Org,
		Name:         req.Name,
		AllowedTools: req.AllowedTools,
		HITLEnabled:  req.HITLEnabled,
	})
	if err != nil {
		http.Error(w, "failed to create profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(profile)
}

func (h *MCPProfileHandler) ListProfiles(w http.ResponseWriter, r *http.Request) {
	org := chi.URLParam(r, "org")
	if org == "" {
		http.Error(w, "org parameter is required", http.StatusBadRequest)
		return
	}

	profiles, err := h.Store.ListAgentProfiles(r.Context(), org)
	if err != nil {
		http.Error(w, "failed to list profiles", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profiles)
}

func (h *MCPProfileHandler) DeleteProfile(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid profile id", http.StatusBadRequest)
		return
	}

	if err := h.Store.DeleteAgentProfile(r.Context(), id); err != nil {
		http.Error(w, "failed to delete profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *MCPProfileHandler) ListHITLQueue(w http.ResponseWriter, r *http.Request) {
	org := chi.URLParam(r, "org")
	if org == "" {
		http.Error(w, "org parameter is required", http.StatusBadRequest)
		return
	}

	items, err := h.Store.ListHITLQueue(r.Context(), org)
	if err != nil {
		http.Error(w, "failed to list HITL queue", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *MCPProfileHandler) ResolveHITLItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid HITL item id", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"` // approved | rejected
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Status != "approved" && req.Status != "rejected" {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	if err := h.Store.ResolveHITLQueueItem(r.Context(), id, req.Status); err != nil {
		http.Error(w, "failed to resolve HITL item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
