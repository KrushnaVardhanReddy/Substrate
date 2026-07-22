package engine

import (
	"encoding/json"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/diff"
)

func DiffSchemas(baseStr, revisionStr string) string {
	rep, err := diff.CompareOpenAPIFromData([]byte(baseStr), []byte(revisionStr), true, nil)
	if err != nil {
		errBytes, _ := json.Marshal(map[string]any{"error": err.Error()})
		return string(errBytes)
	}

	repBytes, err := json.Marshal(rep)
	if err != nil {
		errBytes, _ := json.Marshal(map[string]any{"error": "failed to marshal report: " + err.Error()})
		return string(errBytes)
	}

	return string(repBytes)
}
