package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/cel-go/cel"
)

type CELValidateRequest struct {
	Rule string `json:"rule"`
}

func ValidateCELRuleHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CELValidateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(req.Rule) == "" {
			http.Error(w, "rule cannot be empty", http.StatusBadRequest)
			return
		}

		env, err := cel.NewEnv(
			cel.Variable("request.path", cel.StringType),
			cel.Variable("request.headers", cel.MapType(cel.StringType, cel.StringType)),
			cel.Variable("request.method", cel.StringType),
			cel.Variable("request.auth.claims.group", cel.StringType),
		)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to create CEL env: %v", err), http.StatusInternalServerError)
			return
		}

		_, iss := env.Parse(req.Rule)
		if iss.Err() != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{"valid": false, "error": iss.Err().Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"valid": true})
	}
}
