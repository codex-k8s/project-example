package ws

import (
	"context"
	"time"

	"github.com/codex-k8s/project-example/libs/go/common/logger"
	"github.com/codex-k8s/project-example/services/external/chat-gateway/internal/domain/service"
	"log/slog"
)

// ForwardEvents fans out events from the internal messages service (via Chat.Subscribe)
// to connected WebSocket clients.
//
// This is a transport bridge: domain logic stays in domain/service; here we only orchestrate delivery.
func ForwardEvents(ctx context.Context, log *slog.Logger, chat *service.Chat, hub *Hub) {
	backoff := 1 * time.Second
	for {
		if ctx.Err() != nil {
			return
		}
		ch, err := chat.Subscribe(ctx)
		if err != nil {
			logger.WithContext(ctx, log).Error("subscribe failed", "err", err)
			time.Sleep(backoff)
			continue
		}
		for evt := range ch {
			b, err := EncodeEvent(evt)
			if err != nil {
				logger.WithContext(ctx, log).Warn("encode ws event failed", "err", err)
				continue
			}
			hub.Broadcast(ctx, b)
		}
		// Stream ended; retry with backoff.
		time.Sleep(backoff)
	}
}
