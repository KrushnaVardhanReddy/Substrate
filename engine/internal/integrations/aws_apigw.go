package integrations

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
)

type APIGatewayAPI interface {
	PutRestApi(ctx context.Context, params *apigateway.PutRestApiInput, optFns ...func(*apigateway.Options)) (*apigateway.PutRestApiOutput, error)
}

type AWSAPIGatewayClient struct {
	client APIGatewayAPI
}

func NewAWSAPIGatewayClient(ctx context.Context, region string) (*AWSAPIGatewayClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS config: %w", err)
	}

	client := apigateway.NewFromConfig(cfg)
	return &AWSAPIGatewayClient{client: client}, nil
}

func (c *AWSAPIGatewayClient) PutRestApi(ctx context.Context, restApiId string, spec []byte) error {
	input := &apigateway.PutRestApiInput{
		RestApiId: &restApiId,
		Body:      spec,
		Mode:      "overwrite",
	}

	_, err := c.client.PutRestApi(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to put rest api: %w", err)
	}

	return nil
}
