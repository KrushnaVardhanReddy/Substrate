package main

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
)

func main() {
    loader := openapi3.NewLoader()
    fmt.Printf("%T\n", loader.LoadFromData)
}
