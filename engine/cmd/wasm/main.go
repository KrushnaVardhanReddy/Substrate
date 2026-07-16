//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"

	"github.com/KrushnaVardhanReddy/substrate/engine"
)

func diffSchemasWrapper(this js.Value, args []js.Value) any {
	if len(args) != 2 {
		errBytes, _ := json.Marshal(map[string]any{"error": "Invalid number of arguments, expected 2 (base, revision)"})
		return string(errBytes)
	}

	baseStr := args[0].String()
	revisionStr := args[1].String()

	return engine.DiffSchemas(baseStr, revisionStr)
}

func main() {
	fmt.Println("WASM Engine Initialized")
	js.Global().Set("diff_schemas", js.FuncOf(diffSchemasWrapper))
	// Block indefinitely to keep the WASM instance alive
	<-make(chan struct{})
}
