package service

import (
	"context"
	"time"
)

type UsersAPI interface {
	Register(ctx context.Context, username, password string) (User, error)
	Authenticate(ctx context.Context, username, password string) (User, error)
	GetUser(ctx context.Context, id int64) (User, error)
}

type MessagesAPI interface {
	Create(ctx context.Context, userID int64, text string) (Message, error)
	Delete(ctx context.Context, userID, messageID int64) error
	List(ctx context.Context, limit int) ([]Message, error)
	Subscribe(ctx context.Context) (<-chan Event, error)
}

type Sessions interface {
	Create(ctx context.Context, userID int64, ttl time.Duration) (string, error)
	GetUserID(ctx context.Context, token string) (int64, error)
	Delete(ctx context.Context, token string) error
}
