package webhook

import (
	"encoding/json"
	"net/http"
)

func EnterpriseWebhookPingHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Mock dispatch behavior by accepting the request and returning 200 OK.
		// In a real implementation this would queue a job to push to egress.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "dispatched"})
	}
}
