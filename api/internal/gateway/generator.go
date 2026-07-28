package gateway

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
)

type Generator interface {
	GenerateCRD(spec *openapi3.T, repo string) (string, error)
}

func NewGenerator(gatewayType string) (Generator, error) {
	switch gatewayType {
	case "kong":
		return &KongGenerator{}, nil
	case "aws":
		return &AWSGenerator{}, nil
	default:
		return nil, fmt.Errorf("unsupported gateway type: %s", gatewayType)
	}
}
