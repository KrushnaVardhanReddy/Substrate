// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package integrations

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/apigateway"
)

type mockAPIGatewayAPI struct {
	mockPutRestApi func(ctx context.Context, params *apigateway.PutRestApiInput, optFns ...func(*apigateway.Options)) (*apigateway.PutRestApiOutput, error)
}

func (m mockAPIGatewayAPI) PutRestApi(ctx context.Context, params *apigateway.PutRestApiInput, optFns ...func(*apigateway.Options)) (*apigateway.PutRestApiOutput, error) {
	return m.mockPutRestApi(ctx, params, optFns...)
}

func TestAWSAPIGatewayClient(t *testing.T) {
	ctx := context.Background()
	_, err := NewAWSAPIGatewayClient(ctx, "us-east-1")
	if err != nil {
		t.Fatalf("unexpected error creating client: %v", err)
	}
}

func TestAWSAPIGatewayClient_PutRestApi(t *testing.T) {
	mockClient := mockAPIGatewayAPI{
		mockPutRestApi: func(ctx context.Context, params *apigateway.PutRestApiInput, optFns ...func(*apigateway.Options)) (*apigateway.PutRestApiOutput, error) {
			if *params.RestApiId != "test-api" {
				return nil, fmt.Errorf("unexpected RestApiId")
			}
			return &apigateway.PutRestApiOutput{}, nil
		},
	}

	client := &AWSAPIGatewayClient{client: mockClient}
	err := client.PutRestApi(context.Background(), "test-api", []byte(`{"openapi": "3.0.0"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAWSAPIGatewayClient_PutRestApi_Error(t *testing.T) {
	mockClient := mockAPIGatewayAPI{
		mockPutRestApi: func(ctx context.Context, params *apigateway.PutRestApiInput, optFns ...func(*apigateway.Options)) (*apigateway.PutRestApiOutput, error) {
			return nil, fmt.Errorf("api error")
		},
	}

	client := &AWSAPIGatewayClient{client: mockClient}
	err := client.PutRestApi(context.Background(), "test-api", []byte(`{"openapi": "3.0.0"}`))
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
