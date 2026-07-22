package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
	"unsafe"

	"github.com/KrushnaVardhanReddy/substrate/api/handlers"
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

func getUnexportedChan(obj *handlers.SSEBroker, fieldName string) chan chan string {
	v := reflect.ValueOf(obj).Elem().FieldByName(fieldName)
	ptr := (*chan chan string)(unsafe.Pointer(v.UnsafeAddr()))
	return *ptr
}

func getUnexportedStringChan(obj *handlers.SSEBroker, fieldName string) chan string {
	v := reflect.ValueOf(obj).Elem().FieldByName(fieldName)
	ptr := (*chan string)(unsafe.Pointer(v.UnsafeAddr()))
	return *ptr
}

func TestSSEBroker_100ConcurrentClients(t *testing.T) {
	t.Parallel()
	broker := handlers.NewSSEBroker()
	go broker.Start()
	time.Sleep(50 * time.Millisecond) // allow Start() to initialise

	newClients := getUnexportedChan(broker, "newClients")

	// Register 100 client channels
	for i := 0; i < 100; i++ {
		ch := make(chan string, 1)
		newClients <- ch
	}

	time.Sleep(100 * time.Millisecond) // allow broker to process
	assert.Equal(t, 100, broker.Len())
}

func TestSSEBroker_DisconnectNoLeak(t *testing.T) {
	t.Parallel()
	broker := handlers.NewSSEBroker()
	go broker.Start()
	time.Sleep(50 * time.Millisecond)

	newClients := getUnexportedChan(broker, "newClients")
	defunctClients := getUnexportedChan(broker, "defunctClients")

	// Record number of goroutines before test
	initialGoroutines := runtime.NumGoroutine()

	// Register 100 clients
	clients := make([]chan string, 100)
	for i := 0; i < 100; i++ {
		ch := make(chan string, 1)
		clients[i] = ch
		newClients <- ch
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 100, broker.Len())

	// Disconnect all 100
	for _, ch := range clients {
		defunctClients <- ch
	}

	time.Sleep(500 * time.Millisecond)
	assert.Equal(t, 0, broker.Len())

	// Check for goroutine leak (allow a small margin for system goroutines)
	finalGoroutines := runtime.NumGoroutine()
	assert.LessOrEqual(t, finalGoroutines, initialGoroutines+2, "Expected no significant goroutine leak")
}

func TestSSEBroker_BroadcastUnderConcurrentDisconnect(t *testing.T) {
	t.Parallel()
	broker := handlers.NewSSEBroker()
	go broker.Start()
	time.Sleep(50 * time.Millisecond)

	newClients := getUnexportedChan(broker, "newClients")
	defunctClients := getUnexportedChan(broker, "defunctClients")
	messages := getUnexportedStringChan(broker, "messages")

	// Register 50 clients
	clients := make([]chan string, 50)
	for i := 0; i < 50; i++ {
		ch := make(chan string, 1)
		clients[i] = ch
		newClients <- ch
	}
	time.Sleep(100 * time.Millisecond)
	assert.Equal(t, 50, broker.Len())

	var wg sync.WaitGroup
	wg.Add(3)

	// Goroutine 1: disconnect clients[0:25]
	go func() {
		defer wg.Done()
		for i := 0; i < 25; i++ {
			defunctClients <- clients[i]
		}
	}()

	// Goroutine 2: broadcast a message
	go func() {
		defer wg.Done()
		messages <- "test-event"
	}()

	// Goroutine 3: broadcast another message
	go func() {
		defer wg.Done()
		messages <- "test-event-2"
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("deadlock in TestSSEBroker_BroadcastUnderConcurrentDisconnect")
	}
}

func TestSSEBroker_HeartbeatTick(t *testing.T) {
	t.Parallel()
	broker := handlers.NewSSEBroker()
	go broker.Start()
	time.Sleep(50 * time.Millisecond)

	// Start ServeHTTP using httptest recorder to catch heartbeat without a real server
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/events", nil)

	ctx, cancel := context.WithCancel(context.Background())
	req = req.WithContext(ctx)

	done := make(chan struct{})
	go func() {
		broker.ServeHTTP(recorder, req)
		close(done)
	}()

	// Wait 6 seconds (heartbeat fires every 5s per spec)
	time.Sleep(6 * time.Second)

	cancel() // shut down ServeHTTP goroutine
	<-done   // ensure ServeHTTP has fully exited to avoid race on recorder

	res := recorder.Result()

	buf := new(bytes.Buffer)
	buf.ReadFrom(res.Body)
	received := buf.String()

	assert.True(t, strings.Contains(received, "heartbeat"), "Expected to receive heartbeat")
}
