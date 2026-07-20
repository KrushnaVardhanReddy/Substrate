package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
)

func RegisterResources(server *Server, store db.Store) {
	server.RegisterResource(Resource{
		URI:         "substrate://schemas/{org}/{repo}",
		Name:        "Schema Contracts",
		Description: "Access schema contracts by organization and repository.",
		MimeType:    "application/json",
		Handler: func(uri string) (string, error) {
			prefix := "substrate://schemas/"
			if !strings.HasPrefix(uri, prefix) {
				return "", fmt.Errorf("invalid uri format")
			}
			parts := strings.Split(strings.TrimPrefix(uri, prefix), "/")
			if len(parts) != 2 {
				return "", fmt.Errorf("expected format {org}/{repo}")
			}
			providerFullName := parts[0] + "/" + parts[1]

			ctx := context.Background()
			contracts, err := store.GetContractsByProviderFullName(ctx, providerFullName)
			if err != nil {
				return "", err
			}

			b, err := json.Marshal(contracts)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
	})

	server.RegisterResource(Resource{
		URI:         "substrate://consumers/{org}/{repo}",
		Name:        "Consumer Contracts",
		Description: "Access consumer dependency contracts by provider organization and repository.",
		MimeType:    "application/json",
		Handler: func(uri string) (string, error) {
			prefix := "substrate://consumers/"
			if !strings.HasPrefix(uri, prefix) {
				return "", fmt.Errorf("invalid uri format")
			}
			parts := strings.Split(strings.TrimPrefix(uri, prefix), "/")
			if len(parts) != 2 {
				return "", fmt.Errorf("expected format {org}/{repo}")
			}
			providerFullName := parts[0] + "/" + parts[1]

			ctx := context.Background()
			contracts, err := store.GetContractsByProviderFullName(ctx, providerFullName)
			if err != nil {
				return "", err
			}
			if len(contracts) == 0 {
				return "[]", nil
			}

			consumers, err := store.GetConsumersByProviderContract(ctx, contracts[0].ID)
			if err != nil {
				return "", err
			}

			b, err := json.Marshal(consumers)
			if err != nil {
				return "", err
			}
			return string(b), nil
		},
	})

	server.RegisterResource(Resource{
		URI:         "substrate://governance-rules",
		Name:        "Governance Rules",
		Description: "Access global governance rules for the organization.",
		MimeType:    "application/json",
		Handler: func(uri string) (string, error) {
			return `{"rules": []}`, nil
		},
	})
}

func RegisterPrompts(server *Server) {
	server.RegisterPrompt(Prompt{
		Name:        "substrate_onboarding",
		Description: "Instructions for writing a .substrate-consumer.yaml file",
		Handler: func() (string, error) {
			return "To write a .substrate-consumer.yaml file, specify the provider organization and repository, along with the required operations and fields your service depends on.", nil
		},
	})
}
