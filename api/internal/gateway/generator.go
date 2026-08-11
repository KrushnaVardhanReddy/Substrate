// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

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
