package handlers

import (
	"net/http"
	"os"
)

type AIConfig struct {
	BaseURL string // SUBSTRATE_AI_BASE_URL env var
	APIKey  string // SUBSTRATE_AI_API_KEY env var
	Model   string // SUBSTRATE_AI_MODEL env var
}

func loadAIConfig() AIConfig {
	return AIConfig{
		BaseURL: os.Getenv("SUBSTRATE_AI_BASE_URL"),
		APIKey:  os.Getenv("SUBSTRATE_AI_API_KEY"),
		Model:   os.Getenv("SUBSTRATE_AI_MODEL"),
	}
}

// AIAnalyzeHandler is a stub to allow the code to compile since the full implementation
// is part of P4-T02 but the file was missing in this branch.
func AIAnalyzeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotImplemented)
	}
}
