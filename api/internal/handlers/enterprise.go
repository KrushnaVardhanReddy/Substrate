// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type CELRuleRequest struct {
	Rule string `json:"rule"`
}

type DriftReport struct {
	Status string `json:"status"`
	Report string `json:"report"`
}

func EnterpriseWebhookHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"dispatched"}`))
	}
}

func EnterpriseRulesValidateHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CELRuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		// Mock pass for any rule for E2E purposes unless empty
		if req.Rule == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"valid":false,"error":"rule cannot be empty"}`))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"valid":true}`))
	}
}

func EnterpriseDriftHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := chi.URLParam(r, "org")
		repo := chi.URLParam(r, "repo")

		if org == "" || repo == "" {
			http.Error(w, "missing org or repo", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		resp := DriftReport{
			Status: "drift_detected",
			Report: "Mock sidecar drift report for " + org + "/" + repo,
		}
		json.NewEncoder(w).Encode(resp)
	}
}
