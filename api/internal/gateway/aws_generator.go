// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

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
