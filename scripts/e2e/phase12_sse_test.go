package main

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPhase12SSEResilience(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E tests in short mode")
	}

	waitForP8Services(t)

	// Number of concurrent clients
	numClients := 100
	var wg sync.WaitGroup
	wg.Add(numClients)

	client := &http.Client{
		Timeout: 0, // No timeout for SSE connection
	}

	// 1. Connect 100 clients to /api/v1/events
	for i := 0; i < numClients; i++ {
		go func() {
			defer wg.Done()

			// Abruptly cancel context after 500ms
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			req, err := http.NewRequestWithContext(ctx, "GET", p8ApiURL+"/api/v1/events", nil)
			if err != nil {
				return
			}
			req.Header.Set("Authorization", "Bearer "+createJWT("acme", "admin"))

			resp, err := client.Do(req)
			if err != nil {
				// We expect a context canceled error eventually
				return
			}
			defer resp.Body.Close()

			// Keep reading until context cancels
			buf := make([]byte, 1024)
			for {
				_, err := resp.Body.Read(buf)
				if err != nil {
					break
				}
			}
		}()
	}

	// 2. Wait for all clients to disconnect
	wg.Wait()

	// Give the server a small moment to unregister everything
	time.Sleep(500 * time.Millisecond)

	// 3. Assert the server recovers gracefully
	// We can check if it's responsive to health checks
	healthReq, _ := http.NewRequest("GET", p8ApiURL+"/health", nil)
	healthResp, err := client.Do(healthReq)
	require.NoError(t, err)
	defer healthResp.Body.Close()

	assert.Equal(t, http.StatusOK, healthResp.StatusCode, "Server should be healthy and responsive after SSE stress test")
}
