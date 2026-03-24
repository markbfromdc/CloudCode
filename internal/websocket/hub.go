// Package websocket manages WebSocket connections between browser clients and workspace containers.
// It implements a hub-and-spoke pattern where a central Hub manages all active connections,
// and each connection handles bi-directional data flow with heartbeat monitoring.
package websocket

import (
	"sync"

	"github.com/markbfromdc/cloudcode/internal/logging"
)

// Hub maintains the set of active WebSocket clients and broadcasts
// messages to the appropriate workspace containers.
type Hub struct {
	// clients maps session IDs to their active client connections.
	clients map[string]*Client

	// register is a channel for new client connection requests.
	register chan *Client

	// unregister is a channel for client disconnection requests.
	unregister chan *Client

	mu  sync.RWMutex
	log *logging.Logger
}

// NewHub creates a new Hub instance ready to manage WebSocket connections.
func NewHub(log *logging.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client, 64),
		unregister: make(chan *Client, 64),
		log:        log.WithField("component", "websocket-hub"),
	}
}

// Run starts the Hub's main event loop. This must be called in a goroutine.
// It processes client registrations and unregistrations.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.SessionID] = client
			h.mu.Unlock()
			h.log.Info("client registered: session=%s", client.SessionID)

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.SessionID]; ok {
				delete(h.clients, client.SessionID)
				close(client.send)
			}
			h.mu.Unlock()
			h.log.Info("client unregistered: session=%s", client.SessionID)
		}
	}
}

// Register adds a new client to the hub.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(client *Client) {
	h.unregister <- client
}

// GetClient returns the client for a given session ID, if it exists.
func (h *Hub) GetClient(sessionID string) (*Client, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	c, ok := h.clients[sessionID]
	return c, ok
}

// ActiveSessions returns the number of currently connected clients.
func (h *Hub) ActiveSessions() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
