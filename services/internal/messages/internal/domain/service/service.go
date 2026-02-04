package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/codex-k8s/project-example/services/internal/messages/internal/domain/errs"
	msgrepo "github.com/codex-k8s/project-example/services/internal/messages/internal/domain/repository/message"
	"github.com/codex-k8s/project-example/services/internal/messages/internal/domain/types/entity"
)

// EventType is a stable event name emitted by the messages service.
type EventType string

const (
	// EventMessageCreated indicates a newly created message.
	EventMessageCreated EventType = "message.created"
	// EventMessageDeleted indicates a message has been deleted (soft-delete).
	EventMessageDeleted EventType = "message.deleted"
)

// Event is a real-time event emitted by the messages service.
type Event struct {
	Type      EventType
	Message   *entity.Message
	MessageID int64
	DeletedAt *time.Time
}

// Service implements messages use-cases and exposes an in-process pub/sub for events.
type Service struct {
	repo msgrepo.Repository

	mu   sync.RWMutex
	subs map[chan Event]struct{}
}

// New constructs Service.
func New(repo msgrepo.Repository) *Service {
	return &Service{repo: repo, subs: make(map[chan Event]struct{})}
}

// CreateMessage creates a new message.
func (s *Service) CreateMessage(ctx context.Context, userID int64, text string) (entity.Message, error) {
	if userID <= 0 {
		return entity.Message{}, errs.Validation{Field: "user_id", Msg: "invalid"}
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return entity.Message{}, errs.Validation{Field: "text", Msg: "required"}
	}
	if len(text) > 2000 {
		return entity.Message{}, errs.Validation{Field: "text", Msg: "too long"}
	}

	msg, err := s.repo.Create(ctx, entity.Message{UserID: userID, Text: text})
	if err != nil {
		return entity.Message{}, fmt.Errorf("create message: %w", err)
	}

	s.publish(Event{Type: EventMessageCreated, Message: &msg})
	return msg, nil
}

// DeleteMessage soft-deletes a message if userID is the owner.
func (s *Service) DeleteMessage(ctx context.Context, userID, messageID int64) (entity.Message, error) {
	if userID <= 0 {
		return entity.Message{}, errs.Validation{Field: "user_id", Msg: "invalid"}
	}
	if messageID <= 0 {
		return entity.Message{}, errs.Validation{Field: "message_id", Msg: "invalid"}
	}

	msg, err := s.repo.SoftDelete(ctx, userID, messageID)
	if err != nil {
		var nf errs.NotFound
		if errors.As(err, &nf) {
			return entity.Message{}, nf
		}
		var f errs.Forbidden
		if errors.As(err, &f) {
			return entity.Message{}, f
		}
		return entity.Message{}, fmt.Errorf("delete message: %w", err)
	}

	s.publish(Event{Type: EventMessageDeleted, MessageID: messageID, DeletedAt: msg.DeletedAt})
	return msg, nil
}

// ListRecent returns recent messages with a bounded limit.
func (s *Service) ListRecent(ctx context.Context, limit int) ([]entity.Message, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	out, err := s.repo.ListRecent(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}
	return out, nil
}

// PurgeOld soft-deletes messages created before olderThan.
func (s *Service) PurgeOld(ctx context.Context, olderThan time.Time) ([]entity.Message, error) {
	if olderThan.IsZero() {
		return nil, errs.Validation{Field: "older_than", Msg: "required"}
	}
	out, err := s.repo.PurgeOld(ctx, olderThan)
	if err != nil {
		return nil, fmt.Errorf("purge old messages: %w", err)
	}
	for i := range out {
		msg := out[i]
		s.publish(Event{Type: EventMessageDeleted, MessageID: msg.ID, DeletedAt: msg.DeletedAt})
	}
	return out, nil
}

// Subscribe returns a channel of events; it is closed when ctx is cancelled.
func (s *Service) Subscribe(ctx context.Context) <-chan Event {
	ch := make(chan Event, 64)
	s.mu.Lock()
	s.subs[ch] = struct{}{}
	s.mu.Unlock()

	go func() {
		<-ctx.Done()
		s.mu.Lock()
		delete(s.subs, ch)
		s.mu.Unlock()
		close(ch)
	}()

	return ch
}

func (s *Service) publish(evt Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for ch := range s.subs {
		select {
		case ch <- evt:
		default:
			// Do not block domain operations on slow subscribers.
		}
	}
}
