// SPDX-License-Identifier: MIT
// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// Part of the Substrate diff engine — open-source under the MIT License.
// See engine/LICENSE for details.

package watcher

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// A mock extractor override just for testing so we can count calls
var testExtractCalls int
var extractCallback func() error

func mockExtractSpec() error {
	testExtractCalls++
	if extractCallback != nil {
		return extractCallback()
	}
	return nil
}

// In the real code we call `extractor.ExtractSpec()`. For testing, we can redefine
// the watcher logic or we can just test the debouncer indirectly by mocking the FS
// but we don't have dependency injection for `ExtractSpec`. Let's create a wrapper
// or just modify the test to test the debouncer behavior.
// Actually, modifying `Watcher` to accept a callback is better. Let's patch `watcher.go` slightly.
// But for now, we'll write a table-driven test and patch watcher if needed.

func TestWatcherDebounce(t *testing.T) {
	tests := []struct {
		name          string
		events        int
		delay         time.Duration
		expectedCalls int
	}{
		{
			name:          "single event triggers one call",
			events:        1,
			delay:         0,
			expectedCalls: 1,
		},
		{
			name:          "multiple rapid events trigger one call",
			events:        5,
			delay:         10 * time.Millisecond,
			expectedCalls: 1, // should be 1 because of debouncing (500ms)
		},
		{
			name:          "events spaced out trigger multiple calls",
			events:        2,
			delay:         600 * time.Millisecond,
			expectedCalls: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			calls := 0
			opts := WatchOptions{
				Dir:      dir,
				Debounce: 200 * time.Millisecond, // use 200ms for faster tests
				Logf: func(format string, v ...any) {
					// Ignore logs
				},
				// We need to intercept the extractor call.
				// Let's add ExtractFunc to WatchOptions for dependency injection.
			}

			// We need to wait for the watcher to start
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			started := make(chan struct{})
			done := make(chan struct{})

			// Actually, let's inject a ExtractFunc in WatchOptions
			opts.ExtractFunc = func(filePath string) error {
				calls++
				return nil
			}

			go func() {
				close(started)
				_ = Watch(ctx, opts)
				close(done)
			}()

			<-started
			// give watcher time to set up
			time.Sleep(50 * time.Millisecond)

			for i := 0; i < tt.events; i++ {
				filename := filepath.Join(dir, "test.go")
				_ = os.WriteFile(filename, []byte("package main"), 0644)
				time.Sleep(tt.delay)
			}

			// wait for the final debounce window to close
			time.Sleep(opts.Debounce + 100*time.Millisecond)

			cancel()
			<-done

			if calls != tt.expectedCalls {
				t.Errorf("expected %d calls, got %d", tt.expectedCalls, calls)
			}
		})
	}
}
