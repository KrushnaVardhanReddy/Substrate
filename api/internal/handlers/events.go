package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
	}
}

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

type EventRequest struct {
	Org         string    `json:"org"`
	Repo        string    `json:"repo"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	Timestamp   time.Time `json:"timestamp"`
}


type EventResponse struct {
	ID          int32     `json:"id"`
	Org         string    `json:"org"`
	Repo        string    `json:"repo"`
	EventType   string    `json:"event_type"`
	Description string    `json:"description"`
	EventTime   time.Time `json:"event_time"`
	CreatedAt   time.Time `json:"created_at"`
}

func CreateEventHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req EventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if req.EventType != "deployment" && req.EventType != "incident" {
			http.Error(w, "event_type must be deployment or incident", http.StatusBadRequest)
			return
		}

		if req.Org == "" || req.Repo == "" {
			http.Error(w, "org and repo are required", http.StatusBadRequest)
			return
		}

		if req.Timestamp.IsZero() {
			req.Timestamp = time.Now()
		}

		desc := pgtype.Text{String: req.Description, Valid: req.Description != ""}

				event, err := store.InsertEcosystemEvent(r.Context(), sqlcgen.InsertEcosystemEventParams{
			Org:         req.Org,
			Repo:        req.Repo,
			EventType:   req.EventType,
			Description: desc,
			EventTime:   req.Timestamp,
		})
		if err != nil {
			http.Error(w, "failed to insert event", http.StatusInternalServerError)
			return
		}

		response := EventResponse{
			ID:          event.ID,
			Org:         event.Org,
			Repo:        event.Repo,
			EventType:   event.EventType,
			Description: event.Description.String,
			EventTime:   event.EventTime,
			CreatedAt:   event.CreatedAt,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
	}
}

func GetEventsHandler(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		org := chi.URLParam(r, "org")
		if org == "" {
			http.Error(w, "org is required", http.StatusBadRequest)
			return
		}

		// Default to last 30 days if timeframe not specified via query params (optional enhancement)
		now := time.Now()
		startTime := now.AddDate(0, 0, -30)

				events, err := store.GetEcosystemEventsByOrg(r.Context(), sqlcgen.GetEcosystemEventsByOrgParams{
			Org:       org,
			EventTime: startTime,
            EventTime_2: now,
		})
		if err != nil {
			http.Error(w, "failed to fetch events", http.StatusInternalServerError)
			return
		}

        responses := make([]EventResponse, len(events))
        for i, event := range events {
            responses[i] = EventResponse{
                ID:          event.ID,
                Org:         event.Org,
                Repo:        event.Repo,
                EventType:   event.EventType,
                Description: event.Description.String,
                EventTime:   event.EventTime,
                CreatedAt:   event.CreatedAt,
            }
        }

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(responses)
	}
}
