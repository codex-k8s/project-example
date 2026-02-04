package ws

import (
	"context"
	"log/slog"
	"sync"

	"github.com/codex-k8s/project-example/libs/go/common/logger"
)

// Hub tracks active WebSocket clients and broadcasts messages to them.
type Hub struct {
	log *slog.Logger

	mu      sync.RWMutex
	clients map[*Client]struct{}
}

// NewHub constructs Hub.
func NewHub(log *slog.Logger) *Hub {
	return &Hub{log: log, clients: make(map[*Client]struct{})}
}

// Register adds a client to the hub.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

// Unregister removes a client from the hub.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

// Broadcast sends msg to all clients. Slow clients will drop messages.
func (h *Hub) Broadcast(ctx context.Context, msg []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		select {
		case c.send <- msg:
		default:
			logger.WithContext(ctx, h.log).Warn("ws client slow, dropping message")
		}
	}
}
