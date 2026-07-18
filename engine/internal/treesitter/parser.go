package treesitter

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// FindFieldUsages parses the given file and returns a list of 1-indexed line numbers
// where the targetField is accessed (e.g., as a property access, object key, or destructuring).
func FindFieldUsages(filePath string, targetField string) ([]int, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", filePath, err)
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	var lang *sitter.Language
	if ext == ".ts" || ext == ".tsx" {
		lang = typescript.GetLanguage()
	} else if ext == ".js" || ext == ".jsx" {
		lang = javascript.GetLanguage()
	} else {
		// Unsupported language, return empty slice
		return nil, nil
	}

	parser := sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(lang)

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse file %s: %w", filePath, err)
	}
	defer tree.Close()

	root := tree.RootNode()

	var lineNumbers []int
	seen := make(map[int]bool)

	var traverse func(node *sitter.Node)
	traverse = func(node *sitter.Node) {
		if node == nil {
			return
		}

		nodeType := node.Type()

		// `property_identifier` is used for object keys, property access (obj.prop), etc.
		// `shorthand_property_identifier_pattern` is used for object destructuring (const { prop } = obj)
		if nodeType == "property_identifier" || nodeType == "shorthand_property_identifier_pattern" {
			if node.Content(content) == targetField {
				line := int(node.StartPoint().Row) + 1
				if !seen[line] {
					seen[line] = true
					lineNumbers = append(lineNumbers, line)
				}
			}
		}

		for i := 0; i < int(node.ChildCount()); i++ {
			traverse(node.Child(i))
		}
	}

	traverse(root)
	sort.Ints(lineNumbers)

	return lineNumbers, nil
}
