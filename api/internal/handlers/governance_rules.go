// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type GovernanceRuleRequest struct {
	RuleText string `json:"rule_text"`
}

type GovernanceRulesHandler struct {
	store ports.GovernanceStore
}

func NewGovernanceRulesHandler(store ports.GovernanceStore) *GovernanceRulesHandler {
	return &GovernanceRulesHandler{store: store}
}

func (h *GovernanceRulesHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	orgName := chi.URLParam(r, "org")

	orgID, err := h.store.GetOrgIDByName(r.Context(), orgName)
	if err != nil {
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	rules, err := h.store.GetGovernanceRulesByOrg(r.Context(), orgID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rules)
}

func (h *GovernanceRulesHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	orgName := chi.URLParam(r, "org")

	orgID, err := h.store.GetOrgIDByName(r.Context(), orgName)
	if err != nil {
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	var req GovernanceRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.RuleText == "" {
		http.Error(w, "Rule text is required", http.StatusBadRequest)
		return
	}

	id, err := h.store.UpsertGovernanceRule(r.Context(), orgID, req.RuleText)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{"id": id})
}

func (h *GovernanceRulesHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	orgName := chi.URLParam(r, "org")
	ruleIDStr := chi.URLParam(r, "ruleID")

	orgID, err := h.store.GetOrgIDByName(r.Context(), orgName)
	if err != nil {
		http.Error(w, "Organization not found", http.StatusNotFound)
		return
	}

	ruleID, err := uuid.Parse(ruleIDStr)
	if err != nil {
		http.Error(w, "Invalid rule ID", http.StatusBadRequest)
		return
	}

	if err := h.store.DeleteGovernanceRule(r.Context(), ruleID, orgID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
