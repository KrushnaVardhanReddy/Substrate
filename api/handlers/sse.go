package handlers

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// SSEBroker acts as a channel registry to broadcast events to connected clients
type SSEBroker struct {
	mu sync.RWMutex

	// Channels for managing client connections
	clients  map[chan string]bool
	newClients chan chan string
	defunctClients chan chan string

	// Channel for broadcasting messages to all clients
	messages chan string
}

// NewSSEBroker creates a new SSEBroker instance
func NewSSEBroker() *SSEBroker {
	return &SSEBroker{
		clients:        make(map[chan string]bool),
		newClients:     make(chan chan string),
		defunctClients: make(chan chan string),
		messages:       make(chan string),
	}
}

// Start runs the broker's main loop to manage connections and broadcast messages
func (broker *SSEBroker) Start() {
	for {
		select {
		case s := <-broker.newClients:
			broker.mu.Lock()
			broker.clients[s] = true
			broker.mu.Unlock()
		case s := <-broker.defunctClients:
			broker.mu.Lock()
			delete(broker.clients, s)
			broker.mu.Unlock()
			close(s)
		case msg := <-broker.messages:
			broker.mu.RLock()
			for s := range broker.clients {
				select {
				case s <- msg:
				default:
					// If the client channel is full/blocked, we shouldn't block the broker.
				}
			}
			broker.mu.RUnlock()
		}
	}
}

// Len returns the number of currently connected SSE clients.
// Used only for testing.
func (b *SSEBroker) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

// Broadcast sends a message to all connected clients
func (broker *SSEBroker) Broadcast(msg string) {
	broker.messages <- msg
}

// ServeHTTP handles the SSE connection for a client
func (broker *SSEBroker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Make sure that the writer supports flushing.
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	// Set headers for Server-Sent Events
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// Allowing CORS for testing, could be restricted
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a buffered channel for this client so messages aren't dropped if the client is momentarily busy
	messageChan := make(chan string, 100)

	// Register this client with the broker
	broker.newClients <- messageChan

	// Listen to the closing of the http connection
	notify := r.Context().Done()

	// Ticker for heartbeat to keep connection alive (every 5 seconds)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-notify:
			// Connection closed by client
			broker.defunctClients <- messageChan
			return
		case msg := <-messageChan:
			// Send message to client
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		case <-ticker.C:
			// Send heartbeat
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		}
	}
}
