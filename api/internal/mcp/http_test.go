// Copyright (c) 2026 Krushna Vardhan Reddy. All rights reserved.
// PROPRIETARY AND CONFIDENTIAL — Substrate Cloud Platform.
// Unauthorized copying, modification, or distribution is strictly prohibited.

package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHTTPTransport_ServeSSE(t *testing.T) {
	server := NewServer(nil)
	transport := NewHTTPTransport(server)

	req := httptest.NewRequest("GET", "/mcp/sse", nil)
	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		transport.ServeSSE(w, req)
		close(done)
	}()

	// Wait briefly for the handler to start and write the initial event
	time.Sleep(50 * time.Millisecond)

	// Ensure there is at least one active session recorded while running
	transport.mu.RLock()
	assert.Equal(t, 1, len(transport.sessions))
	transport.mu.RUnlock()

	// Stop handler to prevent data race on ResponseRecorder
	cancel()
	<-done

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, "text/event-stream", res.Header.Get("Content-Type"))
}

func TestHTTPTransport_ServeMessages(t *testing.T) {
	server := NewServer(nil)
	// Register a dummy tool to test JSON-RPC message handling
	server.RegisterTool(Tool{
		Name:        "dummy",
		Description: "dummy tool",
		InputSchema: map[string]any{},
		Handler: func(params json.RawMessage) (any, error) {
			return "dummy success", nil
		},
	})
	transport := NewHTTPTransport(server)

	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "tools/call",
		"params": map[string]any{
			"name":      "dummy",
			"arguments": map[string]any{},
		},
	}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/mcp/message", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	transport.ServeMessages(w, req)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	respBody, _ := io.ReadAll(res.Body)
	assert.Contains(t, string(respBody), "dummy success")
}

func TestHTTPTransport_ServeMessages_WithSessionID(t *testing.T) {
	server := NewServer(nil)
	transport := NewHTTPTransport(server)

	// Simulate an active SSE session
	ch := make(chan []byte, 10)
	transport.mu.Lock()
	transport.sessions["test-session-123"] = ch
	transport.mu.Unlock()

	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "initialize",
	}
	bodyBytes, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/mcp/message?sessionId=test-session-123", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	transport.ServeMessages(w, req)

	res := w.Result()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Verify that the JSON-RPC response was pushed to the SSE channel
	select {
	case msg := <-ch:
		assert.Contains(t, string(msg), "protocolVersion")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected message on SSE channel, but got none")
	}
}

// flushRecorder is a wrapper for httptest.ResponseRecorder that implements http.Flusher
type flushRecorder struct {
	*httptest.ResponseRecorder
}

func (f *flushRecorder) Flush() {
	f.ResponseRecorder.Flush()
}

func TestHTTPTransport_ServeSSE_Write(t *testing.T) {
	server := NewServer(nil)
	transport := NewHTTPTransport(server)

	req, _ := http.NewRequest("GET", "/mcp/sse", nil)
	ctx, cancel := context.WithCancel(req.Context())
	req = req.WithContext(ctx)

	recorder := httptest.NewRecorder()
	fr := &flushRecorder{recorder}

	go func() {
		// Cancel the request context shortly after starting
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	transport.ServeSSE(fr, req)

	body := recorder.Body.String()
	assert.True(t, strings.HasPrefix(body, "event: endpoint\n"))
	assert.Contains(t, body, "data: /mcp/messages?sessionId=")
}
