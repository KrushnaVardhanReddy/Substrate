package mcp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

type HTTPTransport struct {
	server   *Server
	sessions map[string]chan []byte
	mu       sync.RWMutex
}

func NewHTTPTransport(server *Server) *HTTPTransport {
	return &HTTPTransport{
		server:   server,
		sessions: make(map[string]chan []byte),
	}
}

func (h *HTTPTransport) ServeSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// Allow CORS if needed, usually handled by middleware but good to be explicit
	w.Header().Set("Access-Control-Allow-Origin", "*")

	sessionID := uuid.New().String()
	ch := make(chan []byte, 100)

	h.mu.Lock()
	h.sessions[sessionID] = ch
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.sessions, sessionID)
		h.mu.Unlock()
	}()

	// Send initial endpoint event
	fmt.Fprintf(w, "event: endpoint\ndata: /mcp/messages?sessionId=%s\n\n", sessionID)
	flusher.Flush()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", string(msg))
			flusher.Flush()
		}
	}
}

func (h *HTTPTransport) ServeMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	responseBytes := h.server.HandleMessage(body)

	// Broadcast over SSE if sessionId is present
	sessionID := r.URL.Query().Get("sessionId")
	if sessionID != "" {
		h.mu.RLock()
		ch, exists := h.sessions[sessionID]
		h.mu.RUnlock()
		if exists {
			select {
			case ch <- responseBytes:
			default:
				// Channel is full, drop or handle error
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Return 200 OK
	w.Write(responseBytes)
}
