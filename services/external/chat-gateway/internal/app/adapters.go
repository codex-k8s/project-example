package app

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codex-k8s/project-example/services/external/chat-gateway/internal/domain/errs"
	domain "github.com/codex-k8s/project-example/services/external/chat-gateway/internal/domain/service"
	msggen "github.com/codex-k8s/project-example/services/external/chat-gateway/internal/transport/grpc/generated/messages/v1"
	usergen "github.com/codex-k8s/project-example/services/external/chat-gateway/internal/transport/grpc/generated/users/v1"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UsersAdapter struct {
	c usergen.UsersServiceClient
}

func NewUsersAdapter(c usergen.UsersServiceClient) *UsersAdapter { return &UsersAdapter{c: c} }

var _ domain.UsersAPI = (*UsersAdapter)(nil)

func (a *UsersAdapter) Register(ctx context.Context, username, password string) (domain.User, error) {
	resp, err := a.c.Register(ctx, &usergen.RegisterRequest{Username: username, Password: password})
	if err != nil {
		return domain.User{}, mapGRPCError(err)
	}
	return fromProtoUser(resp.GetUser()), nil
}

func (a *UsersAdapter) Authenticate(ctx context.Context, username, password string) (domain.User, error) {
	resp, err := a.c.Authenticate(ctx, &usergen.AuthenticateRequest{Username: username, Password: password})
	if err != nil {
		return domain.User{}, mapGRPCError(err)
	}
	return fromProtoUser(resp.GetUser()), nil
}

func (a *UsersAdapter) GetUser(ctx context.Context, id int64) (domain.User, error) {
	resp, err := a.c.GetUser(ctx, &usergen.GetUserRequest{Id: id})
	if err != nil {
		return domain.User{}, mapGRPCError(err)
	}
	return fromProtoUser(resp.GetUser()), nil
}

type MessagesAdapter struct {
	c msggen.MessagesServiceClient
}

func NewMessagesAdapter(c msggen.MessagesServiceClient) *MessagesAdapter {
	return &MessagesAdapter{c: c}
}

var _ domain.MessagesAPI = (*MessagesAdapter)(nil)

func (a *MessagesAdapter) Create(ctx context.Context, userID int64, text string) (domain.Message, error) {
	resp, err := a.c.CreateMessage(ctx, &msggen.CreateMessageRequest{UserId: userID, Text: text})
	if err != nil {
		return domain.Message{}, mapGRPCError(err)
	}
	return fromProtoMessage(resp.GetMessage()), nil
}

func (a *MessagesAdapter) Delete(ctx context.Context, userID, messageID int64) error {
	_, err := a.c.DeleteMessage(ctx, &msggen.DeleteMessageRequest{UserId: userID, MessageId: messageID})
	if err != nil {
		return mapGRPCError(err)
	}
	return nil
}

func (a *MessagesAdapter) List(ctx context.Context, limit int) ([]domain.Message, error) {
	resp, err := a.c.ListMessages(ctx, &msggen.ListMessagesRequest{Limit: int32(limit)})
	if err != nil {
		return nil, mapGRPCError(err)
	}
	out := make([]domain.Message, 0, len(resp.GetMessages()))
	for _, m := range resp.GetMessages() {
		out = append(out, fromProtoMessage(m))
	}
	return out, nil
}

func (a *MessagesAdapter) Subscribe(ctx context.Context) (<-chan domain.Event, error) {
	stream, err := a.c.SubscribeEvents(ctx, &msggen.SubscribeEventsRequest{})
	if err != nil {
		return nil, mapGRPCError(err)
	}

	ch := make(chan domain.Event, 128)
	go func() {
		defer close(ch)
		for {
			evt, err := stream.Recv()
			if err != nil {
				return
			}
			if evt == nil {
				continue
			}
			out, ok := fromProtoEvent(evt)
			if !ok {
				continue
			}
			select {
			case ch <- out:
			case <-ctx.Done():
				return
			}
		}
	}()

	return ch, nil
}

type SessionsAdapter struct {
	rdb *redis.Client
}

func NewSessionsAdapter(rdb *redis.Client) *SessionsAdapter { return &SessionsAdapter{rdb: rdb} }

var _ domain.Sessions = (*SessionsAdapter)(nil)

func (s *SessionsAdapter) Create(ctx context.Context, userID int64, ttl time.Duration) (string, error) {
	if userID <= 0 {
		return "", errs.Validation{Field: "user_id", Msg: "invalid"}
	}
	token, err := randomToken(32)
	if err != nil {
		return "", fmt.Errorf("create session token: %w", err)
	}
	key := sessionKey(token)
	if err := s.rdb.Set(ctx, key, strconv.FormatInt(userID, 10), ttl).Err(); err != nil {
		return "", fmt.Errorf("store session: %w", err)
	}
	return token, nil
}

func (s *SessionsAdapter) GetUserID(ctx context.Context, token string) (int64, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return 0, errs.Unauthorized{Msg: "missing session"}
	}
	v, err := s.rdb.Get(ctx, sessionKey(token)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, errs.Unauthorized{Msg: "invalid session"}
		}
		return 0, fmt.Errorf("load session: %w", err)
	}
	uid, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, errs.Unauthorized{Msg: "invalid session"}
	}
	return uid, nil
}

func (s *SessionsAdapter) Delete(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return errs.Unauthorized{Msg: "missing session"}
	}
	if err := s.rdb.Del(ctx, sessionKey(token)).Err(); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func sessionKey(token string) string { return "sess:" + token }

func randomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func fromProtoUser(u *usergen.User) domain.User {
	if u == nil {
		return domain.User{}
	}
	t := time.Time{}
	if u.CreatedAt != nil {
		t = u.CreatedAt.AsTime()
	}
	return domain.User{ID: u.Id, Username: u.Username, CreatedAt: t}
}

func fromProtoMessage(m *msggen.Message) domain.Message {
	if m == nil {
		return domain.Message{}
	}
	var deletedAt *time.Time
	if m.DeletedAt != nil {
		t := m.DeletedAt.AsTime()
		deletedAt = &t
	}
	created := time.Time{}
	if m.CreatedAt != nil {
		created = m.CreatedAt.AsTime()
	}
	return domain.Message{
		ID:        m.Id,
		UserID:    m.UserId,
		Text:      m.Text,
		CreatedAt: created,
		DeletedAt: deletedAt,
	}
}

func fromProtoEvent(e *msggen.Event) (domain.Event, bool) {
	switch v := e.GetPayload().(type) {
	case *msggen.Event_MessageCreated:
		msg := fromProtoMessage(v.MessageCreated.GetMessage())
		return domain.Event{Type: domain.EventMessageCreated, Message: &msg}, true
	case *msggen.Event_MessageDeleted:
		var deletedAt *time.Time
		if v.MessageDeleted.GetDeletedAt() != nil {
			t := v.MessageDeleted.GetDeletedAt().AsTime()
			deletedAt = &t
		}
		return domain.Event{
			Type:      domain.EventMessageDeleted,
			MessageID: v.MessageDeleted.GetMessageId(),
			DeletedAt: deletedAt,
		}, true
	default:
		return domain.Event{}, false
	}
}

func mapGRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}
	switch st.Code() {
	case codes.InvalidArgument:
		return errs.Validation{Msg: "invalid argument"}
	case codes.Unauthenticated:
		return errs.Unauthorized{Msg: "unauthorized"}
	case codes.PermissionDenied:
		return errs.Forbidden{Msg: "forbidden"}
	case codes.NotFound:
		return errs.NotFound{Msg: "not found"}
	case codes.AlreadyExists, codes.FailedPrecondition, codes.Aborted:
		return errs.Conflict{Msg: "conflict"}
	default:
		return err
	}
}
