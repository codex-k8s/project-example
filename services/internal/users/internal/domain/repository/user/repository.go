package user

import (
	"context"

	"github.com/codex-k8s/project-example/services/internal/users/internal/domain/types/entity"
)

// Repository is a persistence port for users.
type Repository interface {
	Create(ctx context.Context, u entity.User) (entity.User, error)
	GetByUsername(ctx context.Context, username string) (entity.User, error)
	GetByID(ctx context.Context, id int64) (entity.User, error)
}
