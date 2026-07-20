package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/KrushnaVardhanReddy/substrate/engine/watcher"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch source files and auto-update local OpenAPI spec",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir, _ := os.Getwd()

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
			Dir: dir,
		}

		err := watcher.Watch(ctx, opts)
		if err != nil {
			return err
		}

		return nil
	},
}
