package main

import (
	"context"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/KrushnaVardhanReddy/substrate/engine/internal/config"
	"github.com/KrushnaVardhanReddy/substrate/engine/internal/integrations"
)

var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "Manage API Gateway integrations",
}

var gatewaySyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Push local schema to configured API Gateways",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfgPath, _ := cmd.Flags().GetString("config")
		viper.BindPFlags(cmd.Flags())
		cfg, err := config.LoadConfig(cfgPath)
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		if cfg.Gateway == nil {
			log.Println("No gateway integrations configured.")
			return nil
		}

		specPath := cfg.SpecPath
		if specPath == "" {
			return fmt.Errorf("spec_path must be configured in substrate.yaml")
		}

		spec, err := os.ReadFile(specPath)
		if err != nil {
			return fmt.Errorf("failed to read spec file %s: %w", specPath, err)
		}

		ctx := context.Background()

		// AWS
		for _, gw := range cfg.Gateway.AWS {
			log.Printf("Pushing spec to AWS API Gateway: %s/%s", gw.RestAPIID, gw.StageName)
			client, err := integrations.NewAWSAPIGatewayClient(ctx, gw.Region)
			if err != nil {
				return fmt.Errorf("failed to create AWS APIGW client: %w", err)
			}

			err = client.PutRestApi(ctx, gw.RestAPIID, spec)
			if err != nil {
				return fmt.Errorf("failed to push to AWS APIGW: %w", err)
			}
			log.Printf("Successfully updated AWS API Gateway: %s", gw.RestAPIID)
		}

		// Kong
		for _, gw := range cfg.Gateway.Kong {
			log.Printf("Pushing spec to Kong service: %s", gw.ServiceID)
			client := integrations.NewKongClient(gw.AdminURL, gw.Token)

			err = client.PushServiceSpec(ctx, gw.ServiceID, spec)
			if err != nil {
				return fmt.Errorf("failed to push to Kong: %w", err)
			}
			log.Printf("Successfully updated Kong service: %s", gw.ServiceID)
		}

		return nil
	},
}

func init() {
	gatewaySyncCmd.Flags().String("config", "substrate.yaml", "Path to config file")
	viper.BindPFlags(gatewaySyncCmd.Flags())
	gatewayCmd.AddCommand(gatewaySyncCmd)
	// rootCmd is defined in main.go and we add it there or here explicitly if it's exported.
	// We'll export it in main.go by assigning it or we can just keep it as var if it's in the same package.
}
