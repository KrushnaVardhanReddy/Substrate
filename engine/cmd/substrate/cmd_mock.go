// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package main

import (
	"context"
	"fmt"
	"github.com/KrushnaVardhanReddy/substrate/engine/pkg/mockserver"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	viper.BindPFlags(mockCmd.Flags())
}
