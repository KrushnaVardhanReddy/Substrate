package main
import (
	"fmt"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/parser"
)
func main() {
	schema, _ := parser.ParseSchema(&ast.Source{Input: "type User { id: ID! }\nunion SearchResult = User"})
    def := schema.Definitions.ForName("SearchResult")
    for _, t := range def.Types {
		fmt.Println(t)
	}
}
