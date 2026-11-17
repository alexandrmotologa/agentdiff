package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

// SSEEvent represents a message to send to connected web clients.
type SSEEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// SSEBroker coordinates active Server-Sent Events connections.
type SSEBroker struct {
	mu      sync.Mutex
	clients map[chan []byte]bool
}

// NewSSEBroker creates an initialized SSE broker.
func NewSSEBroker() *SSEBroker {
	return &SSEBroker{
		clients: make(map[chan []byte]bool),
	}
}

// HandleSSE serves HTTP SSE connections.
func (b *SSEBroker) HandleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	msgChan := make(chan []byte, 32)
	b.mu.Lock()
	b.clients[msgChan] = true
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.clients, msgChan)
		close(msgChan)
		b.mu.Unlock()
	}()

	// Send initial connection event
	initData, _ := json.Marshal(map[string]string{"status": "connected"})
	fmt.Fprintf(w, "event: connected\ndata: %s\n\n", initData)
	flusher.Flush()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgChan:
			if !ok {
				return
			}
			w.Write(msg)
			flusher.Flush()
		}
	}
}

// Broadcast sends an event and payload to all connected clients.
func (b *SSEBroker) Broadcast(event string, payload interface{}) {
	dataBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	msg := fmt.Sprintf("event: %s\ndata: %s\n\n", event, dataBytes)
	msgBytes := []byte(msg)

	b.mu.Lock()
	defer b.mu.Unlock()

	for ch := range b.clients {
		select {
		case ch <- msgBytes:
		default:
			// Non-blocking drop if channel buffer full
		}
	}
}
