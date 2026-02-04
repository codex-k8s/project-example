package service

import (
	"context"
	"errors"
	"time"

	"github.com/codex-k8s/project-example/services/external/chat-gateway/internal/domain/errs"
)

// Auth provides authentication and session issuance for the gateway.
type Auth struct {
	users    UsersAPI
	sessions Sessions
	ttl      time.Duration
}

// NewAuth constructs Auth.
func NewAuth(users UsersAPI, sessions Sessions, ttl time.Duration) *Auth {
	return &Auth{users: users, sessions: sessions, ttl: ttl}
}

// Register creates a new user and a session token.
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

// Login authenticates a user and returns a fresh session token.
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

// Logout invalidates an existing session token.
func (a *Auth) Logout(ctx context.Context, token string) error {
	if token == "" {
		return errs.Unauthorized{Msg: "missing session"}
	}
	return a.sessions.Delete(ctx, token)
}

// RequireUserID returns a user ID for an existing session token or an Unauthorized error.
func (a *Auth) RequireUserID(ctx context.Context, token string) (int64, error) {
	if token == "" {
		return 0, errs.Unauthorized{Msg: "missing session"}
	}
	uid, err := a.sessions.GetUserID(ctx, token)
	if err != nil {
		// Session is invalid/expired.
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
