package handlers

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type EventBroker struct {
	clients map[chan []byte]bool
	mu      sync.RWMutex
}

var GlobalEventBroker = &EventBroker{
	clients: make(map[chan []byte]bool),
}

func (b *EventBroker) AddClient(c chan []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[c] = true
}

func (b *EventBroker) RemoveClient(c chan []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.clients[c] {
		delete(b.clients, c)
		// Removed close(c) to avoid panics if Broadcast is implemented later
	}
}

// GetClientCount returns the number of active clients (for testing)
func (b *EventBroker) GetClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

func EventsHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	clientChan := make(chan []byte, 10)
	GlobalEventBroker.AddClient(clientChan)
	defer GlobalEventBroker.RemoveClient(clientChan)

	ctx := r.Context()

	// Initial message
	fmt.Fprintf(w, "event: connected\ndata: ok\n\n")
	flusher.Flush()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-clientChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-ticker.C:
			fmt.Fprintf(w, "data: heartbeat\n\n")
			flusher.Flush()
		}
	}
}
