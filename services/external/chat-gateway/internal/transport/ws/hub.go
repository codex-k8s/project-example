package ws

import (
	"context"
	"log/slog"
	"sync"

	"github.com/codex-k8s/project-example/libs/go/common/logger"
)

type Hub struct {
	log *slog.Logger

	mu      sync.RWMutex
	clients map[*Client]struct{}
}

func NewHub(log *slog.Logger) *Hub {
	return &Hub{log: log, clients: make(map[*Client]struct{})}
}

func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
}

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
