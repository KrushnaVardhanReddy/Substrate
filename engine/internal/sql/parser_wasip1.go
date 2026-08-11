// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

//go:build wasip1
// +build wasip1

package sql

import (
	"fmt"
	"os"
)

func ParseSchema(path string) (*SQLSchema, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("sql parse error: failed to read file: %w", err)
	}
	return ParseSchemaFromBytes(bytes)
}

func ParseSchemaFromBytes(src []byte) (*SQLSchema, error) {
	// Not supported in WASM due to CGO dependency of pg_query
	return nil, fmt.Errorf("sql parsing is not supported on WASI architecture")
}
