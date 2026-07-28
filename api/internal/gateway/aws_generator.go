package gateway

import (
	"encoding/json"
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
)

type AWSGenerator struct{}

func (g *AWSGenerator) GenerateCRD(spec *openapi3.T, repo string) (string, error) {
	if spec == nil {
		return "", fmt.Errorf("openapi spec is nil")
	}

	// Clone spec to avoid mutating the original
	specClone := spec // shallow copy

	data, err := json.MarshalIndent(specClone, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}
