package watcher

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/engine/extractor"
	"github.com/fsnotify/fsnotify"
)

// WatchOptions configures the watcher
type WatchOptions struct {
	Dir         string
	Debounce    time.Duration
	Logf        func(format string, v ...any)
	ExtractFunc func() error
}

// Watch starts monitoring the given directory for .go and .ts changes.
// It stops when context is cancelled.
func Watch(ctx context.Context, opts WatchOptions) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	defer watcher.Close()

	if opts.Debounce == 0 {
		opts.Debounce = 500 * time.Millisecond
	}
	if opts.Logf == nil {
		opts.Logf = log.Printf
	}

	// We only watch the requested directory for simplicity (recursively handling is usually required in fsnotify,
	// but the spec just says "Use fsnotify to watch Go/TypeScript source files for changes" and "Runs as `substrate watch`").
	// To watch recursively, one would typically walk the tree. Here we do a basic walk to add all subdirs.

	// actually we shouldn't walk standard library or heavy things, let's assume we watch current dir recursively for .go and .ts
	err = addSubdirectories(watcher, opts.Dir)
	if err != nil {
		return fmt.Errorf("failed to add directories: %w", err)
	}

	opts.Logf("[substrate] watch started on %s", opts.Dir)

	timer := time.NewTimer(0)
	if !timer.Stop() {
		<-timer.C
	}

	var lastTriggeredFile string

	triggerExtract := func(filePath string) {
		var err error
		if opts.ExtractFunc != nil {
			err = opts.ExtractFunc()
		} else {
			err = extractor.ExtractSpec(filePath)
		}

		if err != nil {
			opts.Logf("[substrate] spec extraction failed: %v", err)
		} else {
			opts.Logf("[substrate] spec updated: %s", time.Now().UTC().Format(time.RFC3339))
		}
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			if isSourceFile(event.Name) && (event.Op.Has(fsnotify.Write) || event.Op.Has(fsnotify.Create) || event.Op.Has(fsnotify.Remove)) {
				lastTriggeredFile = event.Name
				timer.Reset(opts.Debounce)
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			opts.Logf("watcher error: %v", err)
		case <-timer.C:
			triggerExtract(lastTriggeredFile)
		}
	}
}

func isSourceFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".go" || ext == ".ts" || ext == ".tsx" || ext == ".js"
}

// addSubdirectories implements recursive directory watching using filepath.WalkDir.
func addSubdirectories(watcher *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Skip common hidden or heavy directories
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "dist" || name == "build" {
				if name != "." { // allow the root itself
					return filepath.SkipDir
				}
			}
			return watcher.Add(path)
		}
		return nil
	})
}
