// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

//go:build js && wasm

package main

import (
	"syscall/js"

	"github.com/KrushnaVardhanReddy/substrate/engine"
)

func substrateDiffWrapper(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return js.ValueOf(`{"error": "expected 2 arguments: baseStr, revisionStr"}`)
	}

	baseStr := args[0].String()
	revisionStr := args[1].String()

	res := engine.DiffSchemas(baseStr, revisionStr)
	return js.ValueOf(res)
}

func main() {
	js.Global().Set("substrateDiff", js.FuncOf(substrateDiffWrapper))
	<-make(chan struct{})
}
