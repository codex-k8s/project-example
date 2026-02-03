package service

import "time"

type User struct {
	ID        int64
	Username  string
	CreatedAt time.Time
}

type Message struct {
	ID        int64
	UserID    int64
	Text      string
	CreatedAt time.Time
	DeletedAt *time.Time
}

type EventType string

const (
	EventMessageCreated EventType = "message.created"
	EventMessageDeleted EventType = "message.deleted"
)

type Event struct {
	Type      EventType
	Message   *Message
	MessageID int64
	DeletedAt *time.Time
}
