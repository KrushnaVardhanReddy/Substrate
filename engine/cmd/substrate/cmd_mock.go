package main

import (
	"context"
	"fmt"
	"github.com/KrushnaVardhanReddy/substrate/engine/pkg/mockserver"
	"github.com/spf13/cobra"
)

var mockTimestamp string
var mockPort int

var mockCmd = &cobra.Command{
	Use:   "mock",
	Short: "Spin up a local mock server using a historical schema",
	RunE: func(cmd *cobra.Command, args []string) error {
		if mockTimestamp == "" {
			return fmt.Errorf("timestamp is required")
		}

		ctx := context.Background()
		return mockserver.StartMockServer(ctx, mockTimestamp, mockPort)
	},
}

func init() {
	mockCmd.Flags().StringVar(&mockTimestamp, "timestamp", "", "Historical timestamp (YYYY-MM-DD)")
	mockCmd.Flags().IntVar(&mockPort, "port", 8081, "Port to run the mock server on")
}
