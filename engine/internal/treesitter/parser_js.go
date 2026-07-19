//go:build js && wasm
// +build js,wasm

package treesitter

import (
	"fmt"
)

// FindFieldUsages parses the given file and returns a list of 1-indexed line numbers
// where the targetField is accessed (e.g., as a property access, object key, or destructuring).
// This is a stub for the wasm architecture which cannot compile go-tree-sitter.
func FindFieldUsages(filePath string, targetField string) ([]int, error) {
	return nil, fmt.Errorf("FindFieldUsages is not supported on wasm due to CGO dependencies")
}
