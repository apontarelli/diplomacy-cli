package websocket

// Alternative implementation using Server-Sent Events (SSE)
// This uses only the standard library but is unidirectional (server -> client)

/*
import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SSE-based real-time updates (standard library only)
func handleSSE(w http.ResponseWriter, r *http.Request) {
	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get game ID from URL
	gameID := r.PathValue("id")

	// Create a channel for this client
	clientChan := make(chan []byte, 10)

	// Register client with hub (similar to WebSocket)
	// hub.RegisterSSEClient(gameID, clientChan)
	// defer hub.UnregisterSSEClient(gameID, clientChan)

	// Send events to client
	for {
		select {
		case message := <-clientChan:
			fmt.Fprintf(w, "data: %s\n\n", message)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case <-r.Context().Done():
			return
		case <-time.After(30 * time.Second):
			// Send keepalive
			fmt.Fprintf(w, ": keepalive\n\n")
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		}
	}
}

// SSE Pros:
// ✅ Standard library only
// ✅ Simpler protocol
// ✅ Automatic reconnection in browsers
// ✅ Works through proxies better
//
// SSE Cons:
// ❌ Unidirectional only (server -> client)
// ❌ No binary data support
// ❌ Limited browser connection pool (6 per domain)
// ❌ No custom protocols/subprotocols
*/
