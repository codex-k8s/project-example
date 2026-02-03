package service

import (
	"context"
	"errors"
	"time"

	"github.com/codex-k8s/project-example/services/external/chat-gateway/internal/domain/errs"
)

type Auth struct {
	users    UsersAPI
	sessions Sessions
	ttl      time.Duration
}

func NewAuth(users UsersAPI, sessions Sessions, ttl time.Duration) *Auth {
	return &Auth{users: users, sessions: sessions, ttl: ttl}
}

func (a *Auth) Register(ctx context.Context, username, password string) (User, string, error) {
	u, err := a.users.Register(ctx, username, password)
	if err != nil {
		return User{}, "", err
	}
	token, err := a.sessions.Create(ctx, u.ID, a.ttl)
	if err != nil {
		return User{}, "", err
	}
	return u, token, nil
}

func (a *Auth) Login(ctx context.Context, username, password string) (User, string, error) {
	u, err := a.users.Authenticate(ctx, username, password)
	if err != nil {
		return User{}, "", err
	}
	token, err := a.sessions.Create(ctx, u.ID, a.ttl)
	if err != nil {
		return User{}, "", err
	}
	return u, token, nil
}

func (a *Auth) Logout(ctx context.Context, token string) error {
	if token == "" {
		return errs.Unauthorized{Msg: "missing session"}
	}
	return a.sessions.Delete(ctx, token)
}

func (a *Auth) RequireUserID(ctx context.Context, token string) (int64, error) {
	if token == "" {
		return 0, errs.Unauthorized{Msg: "missing session"}
	}
	uid, err := a.sessions.GetUserID(ctx, token)
	if err != nil {
		// Сессия невалидна/истекла.
		var u errs.Unauthorized
		if errors.As(err, &u) {
			return 0, u
		}
		return 0, err
	}
	if uid <= 0 {
		return 0, errs.Unauthorized{Msg: "invalid session"}
	}
	return uid, nil
}
