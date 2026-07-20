package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/KrushnaVardhanReddy/substrate/engine/linter"
	"github.com/spf13/cobra"
)

var lintSchemaPath string

var lintCmd = &cobra.Command{
	Use:   "lint",
	Short: "Detect schema smells in API designs",
	RunE: func(cmd *cobra.Command, args []string) error {
		if lintSchemaPath == "" {
			return fmt.Errorf("--schema is required")
		}

		schemaContent, err := os.ReadFile(lintSchemaPath)
		if err != nil {
			return fmt.Errorf("failed to read schema: %w", err)
		}

		score, report, err := linter.Analyze(context.Background(), schemaContent)
		if err != nil {
			return fmt.Errorf("linter analysis failed: %w", err)
		}

		out := map[string]interface{}{
			"score":  score,
			"issues": report,
		}

		b, err := json.MarshalIndent(out, "", "  ")
		if err != nil {
			return err
		}

		fmt.Println(string(b))
		return nil
	},
}

func init() {
	lintCmd.Flags().StringVar(&lintSchemaPath, "schema", "", "Path to the OpenAPI schema")
}
