// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package gateway

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/config"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/github"
	"github.com/getkin/kin-openapi/openapi3"
)

func generatePRBranchName(repo string) string {
	return fmt.Sprintf("substrate/gateway-sync-%s-%d", repo, time.Now().Unix())
}

func RunGatewaySync(ctx context.Context, store db.Store, ghClient github.Client, org, repo string) (string, error) {
	content, err := ghClient.GetFileContent(ctx, org, repo, "substrate.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to get substrate.yaml: %w", err)
	}

	cfg, err := config.Parse([]byte(content))
	if err != nil {
		return "", fmt.Errorf("failed to parse substrate.yaml: %w", err)
	}

	if cfg.Gateway == nil || cfg.Gateway.Type == "" {
		return "", nil
	}

	schemaPath := cfg.HeadSchema
	if schemaPath == "" {
		schemaPath = "openapi.yaml"
	}

	schemaContent, err := ghClient.GetFileContent(ctx, org, repo, schemaPath)
	if err != nil {
		return "", fmt.Errorf("failed to get schema %s: %w", schemaPath, err)
	}

	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromData([]byte(schemaContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse openapi spec: %w", err)
	}

	gen, err := NewGenerator(cfg.Gateway.Type)
	if err != nil {
		return "", fmt.Errorf("failed to create generator: %w", err)
	}

	crdYaml, err := gen.GenerateCRD(doc, repo)
	if err != nil {
		return "", fmt.Errorf("failed to generate crd: %w", err)
	}

	var prUrl string
	if cfg.Gateway.InfraRepo != "" {
		infraRepoParts := bytes.Split([]byte(cfg.Gateway.InfraRepo), []byte("/"))
		if len(infraRepoParts) == 2 {
			infraOwner := string(infraRepoParts[0])
			infraRepoName := string(infraRepoParts[1])

			branch := generatePRBranchName(repo)

			outputPath := cfg.Gateway.OutputPath
			if outputPath == "" {
				outputPath = "gateway/"
			}
			if outputPath[len(outputPath)-1] != '/' {
				outputPath += "/"
			}
			filePath := fmt.Sprintf("%ssubstrate-generated-%s.yaml", outputPath, repo)

			title := fmt.Sprintf("chore(gateway): Sync API Gateway for %s", repo)
			body := fmt.Sprintf("Automatically generated API Gateway configuration for %s/%s.", org, repo)

			err = ghClient.CommitAndPushFile(ctx, infraOwner, infraRepoName, branch, filePath, crdYaml, title)
			if err != nil {
				log.Printf("Failed to commit and push file to infra repo: %v", err)
			} else {
				prUrl, err = ghClient.CreatePR(ctx, infraOwner, infraRepoName, title, branch, body)
				if err != nil {
					log.Printf("Failed to create PR in infra repo: %v", err)
				}
			}
		}
	}

	err = store.SaveGatewayConfig(ctx, org, repo, cfg.Gateway.Type, crdYaml, prUrl)
	if err != nil {
		return "", fmt.Errorf("failed to save gateway config: %w", err)
	}

	return crdYaml, nil
}
