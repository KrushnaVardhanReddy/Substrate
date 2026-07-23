package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/KrushnaVardhanReddy/substrate/engine/extractor"
	"github.com/KrushnaVardhanReddy/substrate/engine/watcher"
	"github.com/spf13/cobra"
)

var watchDir string
var specOutPath string

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch source files and auto-update local OpenAPI spec",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Handle graceful shutdown
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigCh
			fmt.Println("\n[substrate] Stopping watch...")
			cancel()
		}()

		opts := watcher.WatchOptions{
			Dir: watchDir,
			ExtractFunc: func(filePath string) error {
				err := extractor.ExtractSpec(filePath)
				if err != nil {
					return err
				}
				if specOutPath != "" && specOutPath != "openapi.yaml" {
					return os.Rename("openapi.yaml", specOutPath)
				}
				return nil
			},
		}

		err := watcher.Watch(ctx, opts)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	watchCmd.Flags().StringVar(&watchDir, "dir", ".", "Directory to watch")
	watchCmd.Flags().StringVar(&specOutPath, "spec", "openapi.yaml", "Output path for the OpenAPI spec")
}
