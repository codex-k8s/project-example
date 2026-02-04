package http

import (
	"net/http"
	"time"

	"github.com/codex-k8s/project-example/services/external/chat-gateway/internal/domain/service"
	"github.com/codex-k8s/project-example/services/external/chat-gateway/internal/transport/http/generated"
	"github.com/labstack/echo/v5"
)

// Handler implements the generated OpenAPI server interface for chat-gateway.
type Handler struct {
	auth       *service.Auth
	chat       *service.Chat
	cookieName string
	cookieTTL  time.Duration
	secure     bool
}

// NewHandler constructs Handler.
func NewHandler(auth *service.Auth, chat *service.Chat, cookieName string, cookieTTL time.Duration, cookieSecure bool) *Handler {
	return &Handler{
		auth:       auth,
		chat:       chat,
		cookieName: cookieName,
		cookieTTL:  cookieTTL,
		secure:     cookieSecure,
	}
}

var _ generated.ServerInterface = (*Handler)(nil)

// AuthRegister handles POST /auth/register.
func (h *Handler) AuthRegister(c *echo.Context) error {
	var body generated.AuthRegisterJSONRequestBody
	if err := c.Bind(&body); err != nil {
		return err
	}

	u, token, err := h.auth.Register(c.Request().Context(), body.Username, body.Password)
	if err != nil {
		return err
	}

	h.setSessionCookie(c, token)
	return c.JSON(http.StatusCreated, toHTTPUser(u))
}

// AuthLogin handles POST /auth/login.
func (h *Handler) AuthLogin(c *echo.Context) error {
	var body generated.AuthLoginJSONRequestBody
	if err := c.Bind(&body); err != nil {
		return err
	}

	u, token, err := h.auth.Login(c.Request().Context(), body.Username, body.Password)
	if err != nil {
		return err
	}

	h.setSessionCookie(c, token)
	return c.JSON(http.StatusOK, toHTTPUser(u))
}

// AuthLogout handles POST /auth/logout.
func (h *Handler) AuthLogout(c *echo.Context) error {
	token := h.getSessionToken(c)
	if err := h.auth.Logout(c.Request().Context(), token); err != nil {
		return err
	}
	h.clearSessionCookie(c)
	return c.NoContent(http.StatusNoContent)
}

// MessagesList handles GET /messages.
func (h *Handler) MessagesList(c *echo.Context, params generated.MessagesListParams) error {
	limit := 50
	if params.Limit != nil {
		limit = *params.Limit
	}
	msgs, err := h.chat.ListMessages(c.Request().Context(), limit)
	if err != nil {
		return err
	}

	out := make([]generated.Message, 0, len(msgs))
	for i := range msgs {
		out = append(out, toHTTPMessage(msgs[i]))
	}
	return c.JSON(http.StatusOK, generated.ListMessagesResponse{Messages: out})
}

// MessagesCreate handles POST /messages.
func (h *Handler) MessagesCreate(c *echo.Context) error {
	token := h.getSessionToken(c)
	userID, err := h.auth.RequireUserID(c.Request().Context(), token)
	if err != nil {
		return err
	}

	var body generated.MessagesCreateJSONRequestBody
	if err := c.Bind(&body); err != nil {
		return err
	}

	msg, err := h.chat.CreateMessage(c.Request().Context(), userID, body.Text)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, toHTTPMessage(msg))
}

// MessagesDelete handles DELETE /messages/{id}.
func (h *Handler) MessagesDelete(c *echo.Context, id int) error {
	token := h.getSessionToken(c)
	userID, err := h.auth.RequireUserID(c.Request().Context(), token)
	if err != nil {
		return err
	}
	if err := h.chat.DeleteMessage(c.Request().Context(), userID, int64(id)); err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) getSessionToken(c *echo.Context) string {
	ck, err := c.Cookie(h.cookieName)
	if err != nil || ck == nil {
		return ""
	}
	return ck.Value
}

func (h *Handler) setSessionCookie(c *echo.Context, token string) {
	c.SetCookie(&http.Cookie{
		Name:     h.cookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.cookieTTL.Seconds()),
	})
}

func (h *Handler) clearSessionCookie(c *echo.Context) {
	c.SetCookie(&http.Cookie{
		Name:     h.cookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

func toHTTPUser(u service.User) generated.User {
	return generated.User{Id: u.ID, Username: u.Username, CreatedAt: u.CreatedAt}
}

func toHTTPMessage(m service.Message) generated.Message {
	return generated.Message{
		Id:        m.ID,
		UserId:    m.UserID,
		Text:      m.Text,
		CreatedAt: m.CreatedAt,
		DeletedAt: m.DeletedAt,
	}
}
