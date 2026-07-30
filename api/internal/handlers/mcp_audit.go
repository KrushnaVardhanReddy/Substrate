package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

type MCPAuditHandler struct {
	Store db.Store
}

func (h *MCPAuditHandler) ListMCPAuditLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := h.Store.ListMCPAuditLogs(r.Context())
	if err != nil {
		http.Error(w, "failed to get audit logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(logs); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
