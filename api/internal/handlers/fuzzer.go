package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/ports"
)

type FuzzerHandler struct {
	Store ports.Store
}

func (h *FuzzerHandler) GetSchemaValidationGaps(w http.ResponseWriter, r *http.Request) {
	gaps, err := h.Store.GetSchemaValidationGaps(r.Context())
	if err != nil {
		http.Error(w, "Failed to get schema validation gaps", http.StatusInternalServerError)
		return
	}

	type GapResponse struct {
		ID       string `json:"ID"`
		Method   string `json:"Method"`
		Path     string `json:"Path"`
		Payload  json.RawMessage `json:"Payload"`
		Issue    string `json:"Issue"`
		Severity string `json:"Severity"`
	}

	var response []GapResponse
	for _, g := range gaps {
		response = append(response, GapResponse{
			ID:       g.ID.String(),
			Method:   g.Method,
			Path:     g.Path,
			Payload:  g.Payload,
			Issue:    g.Issue,
			Severity: g.Severity,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
