package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/diff"
)

func diffSchemasWrapper(this js.Value, args []js.Value) any {
	if len(args) != 2 {
		return map[string]any{"error": "Invalid number of arguments, expected 2 (base, revision)"}
	}

	baseStr := args[0].String()
	revisionStr := args[1].String()

	rep, err := diff.CompareOpenAPIFromData([]byte(baseStr), []byte(revisionStr), true, nil)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}

	repBytes, err := json.Marshal(rep)
	if err != nil {
		return map[string]any{"error": "failed to marshal report: " + err.Error()}
	}

	return string(repBytes)
}

func main() {
	fmt.Println("WASM Engine Initialized")
	js.Global().Set("diff_schemas", js.FuncOf(diffSchemasWrapper))
	// Block indefinitely to keep the WASM instance alive
	<-make(chan struct{})
}
