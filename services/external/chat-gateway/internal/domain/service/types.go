package service

import "time"

// User is the gateway-level user model used by gateway use-cases.
type User struct {
	ID        int64
	Username  string
	CreatedAt time.Time
}

// Message is the gateway-level message model used by gateway use-cases.
type Message struct {
	ID        int64
	UserID    int64
	Text      string
	CreatedAt time.Time
	DeletedAt *time.Time
}

// EventType is a stable event name delivered to WebSocket clients.
type EventType string

const (
	// EventMessageCreated indicates a newly created message.
	EventMessageCreated EventType = "message.created"
	// EventMessageDeleted indicates a message has been deleted (soft-delete).
	EventMessageDeleted EventType = "message.deleted"
)

// Event is a normalized real-time event.
type Event struct {
	Type      EventType
	Message   *Message
	MessageID int64
	DeletedAt *time.Time
}
