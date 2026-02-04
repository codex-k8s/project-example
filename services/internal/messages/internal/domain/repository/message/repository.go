package message

import (
	"context"
	"time"

	"github.com/codex-k8s/project-example/services/internal/messages/internal/domain/types/entity"
)

// Repository is a persistence port for messages.
type Repository interface {
	Create(ctx context.Context, msg entity.Message) (entity.Message, error)
	SoftDelete(ctx context.Context, userID, messageID int64) (entity.Message, error)
	ListRecent(ctx context.Context, limit int) ([]entity.Message, error)
	PurgeOld(ctx context.Context, olderThan time.Time) ([]entity.Message, error)
}
