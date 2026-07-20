package agents

import (
	"encoding/json"
	"net/http"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

type AgentRegistrationRequest struct {
	RepoName string             `json:"repo_name"`
	Owner    string             `json:"owner"`
	Tools    []AgentToolRequest `json:"tools"`
}

type AgentToolRequest struct {
	ToolName       string          `json:"tool_name"`
	ParametersJSON json.RawMessage `json:"parameters_json"`
}

func RegisterAgentHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req AgentRegistrationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if req.RepoName == "" || req.Owner == "" {
			http.Error(w, "repo_name and owner are required", http.StatusBadRequest)
			return
		}

		var tools []db.AgentToolDependency
		for _, t := range req.Tools {
			if t.ToolName == "" {
				http.Error(w, "tool_name is required for all tools", http.StatusBadRequest)
				return
			}
			tools = append(tools, db.AgentToolDependency{
				ToolName:       t.ToolName,
				ParametersJSON: t.ParametersJSON,
			})
		}

		id, err := store.RegisterAgent(r.Context(), req.RepoName, req.Owner, tools)
		if err != nil {
			http.Error(w, "failed to register agent", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      id.String(),
			"message": "agent registered successfully",
		})
	}
}
