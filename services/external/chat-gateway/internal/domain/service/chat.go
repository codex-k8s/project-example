package service

import "context"

type Chat struct {
	msgs MessagesAPI
}

func NewChat(msgs MessagesAPI) *Chat { return &Chat{msgs: msgs} }

func (c *Chat) CreateMessage(ctx context.Context, userID int64, text string) (Message, error) {
	return c.msgs.Create(ctx, userID, text)
}

func (c *Chat) DeleteMessage(ctx context.Context, userID, messageID int64) error {
	return c.msgs.Delete(ctx, userID, messageID)
}

func (c *Chat) ListMessages(ctx context.Context, limit int) ([]Message, error) {
	return c.msgs.List(ctx, limit)
}

func (c *Chat) Subscribe(ctx context.Context) (<-chan Event, error) {
	return c.msgs.Subscribe(ctx)
}
