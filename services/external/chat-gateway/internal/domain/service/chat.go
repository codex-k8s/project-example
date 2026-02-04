package service

import "context"

// Chat provides chat-related use-cases for the gateway.
type Chat struct {
	msgs MessagesAPI
}

// NewChat constructs Chat.
func NewChat(msgs MessagesAPI) *Chat { return &Chat{msgs: msgs} }

// CreateMessage creates a new message.
func (c *Chat) CreateMessage(ctx context.Context, userID int64, text string) (Message, error) {
	return c.msgs.Create(ctx, userID, text)
}

// DeleteMessage deletes (soft-deletes) a message owned by userID.
func (c *Chat) DeleteMessage(ctx context.Context, userID, messageID int64) error {
	return c.msgs.Delete(ctx, userID, messageID)
}

// ListMessages returns recent messages.
func (c *Chat) ListMessages(ctx context.Context, limit int) ([]Message, error) {
	return c.msgs.List(ctx, limit)
}

// Subscribe returns a channel of real-time events.
func (c *Chat) Subscribe(ctx context.Context) (<-chan Event, error) {
	return c.msgs.Subscribe(ctx)
}
