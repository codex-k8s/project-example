package ws

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/codex-k8s/project-example/services/external/chat-gateway/internal/domain/service"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"
)

type Handler struct {
	hub        *Hub
	auth       *service.Auth
	cookieName string
}

func NewHandler(hub *Hub, auth *service.Auth, cookieName string) *Handler {
	return &Handler{hub: hub, auth: auth, cookieName: cookieName}
}

func (h *Handler) Handle(c *echo.Context) error {
	token := ""
	if ck, err := c.Cookie(h.cookieName); err == nil && ck != nil {
		token = ck.Value
	}
	uid, err := h.auth.RequireUserID(c.Request().Context(), token)
	if err != nil {
		return err
	}

	up := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := strings.TrimSpace(r.Header.Get("Origin"))
			if origin == "" {
				return false
			}
			// По умолчанию: только HTTPS same-origin (клиенты ходят только по https/wss).
			if origin == "https://"+r.Host {
				return true
			}
			// Дополнительно можно разрешить явный allowlist (через env).
			if raw := strings.TrimSpace(os.Getenv("WS_ALLOWED_ORIGINS")); raw != "" {
				for _, v := range strings.Split(raw, ",") {
					if strings.TrimSpace(v) == origin {
						return true
					}
				}
			}
			return false
		},
	}
	conn, err := up.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()

	client := newClient(uid, conn)
	h.hub.Register(client)
	defer h.hub.Unregister(client)

	ctx, cancel := context.WithCancel(c.Request().Context())
	defer cancel()

	go client.writePump(ctx)
	client.readPump(ctx)

	// Дадим writePump возможность отправить close frame.
	_ = conn.SetWriteDeadline(time.Now().Add(1 * time.Second))
	return nil
}
