// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type PartnersHandler struct {
	store ports.Store
}

func NewPartnersHandler(store ports.Store) *PartnersHandler {
	return &PartnersHandler{store: store}
}

type PartnerIntegration struct {
	ID            string    `json:"id"`
	VendorName    string    `json:"vendor_name"`
	Status        string    `json:"status"`
	WebhookURL    string    `json:"webhook_url"`
	WebhookSecret string    `json:"webhook_secret,omitempty"` // Omitted in most responses
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreatePartnerRequest struct {
	VendorName    string `json:"vendor_name"`
	WebhookURL    string `json:"webhook_url"`
	WebhookSecret string `json:"webhook_secret"`
}

type UpdatePartnerRequest struct {
	VendorName    string `json:"vendor_name,omitempty"`
	WebhookURL    string `json:"webhook_url,omitempty"`
	WebhookSecret string `json:"webhook_secret,omitempty"`
}

func convertPartner(p sqlcgen.PartnerIntegration) PartnerIntegration {
	return PartnerIntegration{
		ID:            uuid.UUID(p.ID.Bytes).String(),
		VendorName:    p.VendorName,
		Status:        p.Status,
		WebhookURL:    p.WebhookUrl,
		WebhookSecret: p.WebhookSecret,
		CreatedAt:     p.CreatedAt.Time,
		UpdatedAt:     p.UpdatedAt.Time,
	}
}

func (h *PartnersHandler) ListPartners(w http.ResponseWriter, r *http.Request) {
	rows, err := h.store.ListPartners(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch partners", http.StatusInternalServerError)
		return
	}

	partners := []PartnerIntegration{}
	for _, p := range rows {
		partners = append(partners, convertPartner(p))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(partners)
}

func (h *PartnersHandler) CreatePartner(w http.ResponseWriter, r *http.Request) {
	var req CreatePartnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.VendorName == "" || req.WebhookURL == "" || req.WebhookSecret == "" {
		http.Error(w, "VendorName, WebhookURL, and WebhookSecret are required", http.StatusBadRequest)
		return
	}

	p, err := h.store.CreatePartner(r.Context(), sqlcgen.CreatePartnerParams{
		VendorName:    req.VendorName,
		WebhookUrl:    req.WebhookURL,
		WebhookSecret: req.WebhookSecret,
	})

	if err != nil {
		http.Error(w, "Failed to create partner integration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(convertPartner(p))
}

func (h *PartnersHandler) UpdatePartner(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "Missing partner ID", http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid partner ID", http.StatusBadRequest)
		return
	}

	var req UpdatePartnerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	pgUUID := pgtype.UUID{Bytes: id, Valid: true}

	p, err := h.store.UpdatePartner(r.Context(), sqlcgen.UpdatePartnerParams{
		VendorName:    req.VendorName,
		WebhookUrl:    req.WebhookURL,
		WebhookSecret: req.WebhookSecret,
		ID:            pgUUID,
	})

	if err != nil {
		if err.Error() == "no rows in result set" {
			http.Error(w, "Partner not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update partner", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(convertPartner(p))
}

func (h *PartnersHandler) DeletePartner(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "Missing partner ID", http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid partner ID", http.StatusBadRequest)
		return
	}

	err = h.store.DeletePartner(r.Context(), id)

	if err != nil {
		if err.Error() == "no rows in result set" || err.Error() == "not found" {
			http.Error(w, "Partner not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete partner", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Partner deleted"})
}

func (h *PartnersHandler) VerifyPartner(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "Missing partner ID", http.StatusBadRequest)
		return
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid partner ID", http.StatusBadRequest)
		return
	}

	p, err := h.store.GetPartner(r.Context(), id)

	if err != nil {
		http.Error(w, "Partner not found", http.StatusNotFound)
		return
	}

	// Handshake logic
	challenge := uuid.New().String()
	payload := map[string]string{"challenge": challenge}
	payloadBytes, _ := json.Marshal(payload)

	mac := hmac.New(sha256.New, []byte(p.WebhookSecret))
	mac.Write(payloadBytes)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))

	req, err := http.NewRequestWithContext(r.Context(), "POST", p.WebhookUrl, bytes.NewBuffer(payloadBytes))
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Substrate-Signature", expectedSignature)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)

	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Verification failed: webhook did not return 200 OK", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	pgUUID := pgtype.UUID{Bytes: id, Valid: true}

	// Verification succeeded, update status
	updatedP, err := h.store.UpdatePartnerStatus(r.Context(), sqlcgen.UpdatePartnerStatusParams{
		Status: "certified",
		ID:     pgUUID,
	})

	if err != nil {
		http.Error(w, "Failed to update partner status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(convertPartner(updatedP))
}
