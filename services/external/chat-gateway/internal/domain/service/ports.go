package service

import (
	"context"
	"time"
)

// UsersAPI is a gateway port for the users bounded context.
type UsersAPI interface {
	Register(ctx context.Context, username, password string) (User, error)
	Authenticate(ctx context.Context, username, password string) (User, error)
	GetUser(ctx context.Context, id int64) (User, error)
}

// MessagesAPI is a gateway port for the messages bounded context.
type MessagesAPI interface {
	Create(ctx context.Context, userID int64, text string) (Message, error)
	Delete(ctx context.Context, userID, messageID int64) error
	List(ctx context.Context, limit int) ([]Message, error)
	Subscribe(ctx context.Context) (<-chan Event, error)
}

// Sessions is a gateway port for session storage.
type Sessions interface {
	Create(ctx context.Context, userID int64, ttl time.Duration) (string, error)
	GetUserID(ctx context.Context, token string) (int64, error)
	Delete(ctx context.Context, token string) error
}
