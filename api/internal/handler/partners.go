package handler

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PartnersHandler struct {
	pool *pgxpool.Pool
}

func NewPartnersHandler(pool *pgxpool.Pool) *PartnersHandler {
	return &PartnersHandler{pool: pool}
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

func (h *PartnersHandler) ListPartners(w http.ResponseWriter, r *http.Request) {
	rows, err := h.pool.Query(r.Context(), `
		SELECT id, vendor_name, status, webhook_url, created_at, updated_at
		FROM partner_integrations
		ORDER BY created_at DESC
	`)
	if err != nil {
		http.Error(w, "Failed to fetch partners", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	partners := []PartnerIntegration{}
	for rows.Next() {
		var p PartnerIntegration
		err := rows.Scan(&p.ID, &p.VendorName, &p.Status, &p.WebhookURL, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			http.Error(w, "Failed to scan partner", http.StatusInternalServerError)
			return
		}
		partners = append(partners, p)
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

	var p PartnerIntegration
	err := h.pool.QueryRow(r.Context(), `
		INSERT INTO partner_integrations (vendor_name, webhook_url, webhook_secret)
		VALUES ($1, $2, $3)
		RETURNING id, vendor_name, status, webhook_url, created_at, updated_at
	`, req.VendorName, req.WebhookURL, req.WebhookSecret).Scan(
		&p.ID, &p.VendorName, &p.Status, &p.WebhookURL, &p.CreatedAt, &p.UpdatedAt,
	)

	if err != nil {
		http.Error(w, "Failed to create partner integration", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *PartnersHandler) VerifyPartner(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Missing partner ID", http.StatusBadRequest)
		return
	}

	var p PartnerIntegration
	err := h.pool.QueryRow(r.Context(), `
		SELECT id, webhook_url, webhook_secret, status
		FROM partner_integrations
		WHERE id = $1
	`, id).Scan(&p.ID, &p.WebhookURL, &p.WebhookSecret, &p.Status)

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

	req, err := http.NewRequestWithContext(r.Context(), "POST", p.WebhookURL, bytes.NewBuffer(payloadBytes))
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

	// Verification succeeded, update status
	err = h.pool.QueryRow(r.Context(), `
		UPDATE partner_integrations
		SET status = 'certified', updated_at = NOW()
		WHERE id = $1
		RETURNING status, updated_at
	`, id).Scan(&p.Status, &p.UpdatedAt)

	if err != nil {
		http.Error(w, "Failed to update partner status", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}
