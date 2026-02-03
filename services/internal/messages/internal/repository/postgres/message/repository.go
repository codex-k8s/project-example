package message

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/codex-k8s/project-example/services/internal/messages/internal/domain/errs"
	msgrepo "github.com/codex-k8s/project-example/services/internal/messages/internal/domain/repository/message"
	"github.com/codex-k8s/project-example/services/internal/messages/internal/domain/types/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/create_message.sql
var sqlCreate string

//go:embed sql/get_message_by_id.sql
var sqlGetByID string

//go:embed sql/soft_delete_message.sql
var sqlSoftDelete string

//go:embed sql/list_recent_messages.sql
var sqlListRecent string

//go:embed sql/purge_old_messages.sql
var sqlPurgeOld string

type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

var _ msgrepo.Repository = (*Repo)(nil)

func (r *Repo) Create(ctx context.Context, msg entity.Message) (entity.Message, error) {
	row := r.pool.QueryRow(ctx, sqlCreate, msg.UserID, msg.Text)
	var out entity.Message
	var deletedAt *time.Time
	if err := row.Scan(&out.ID, &out.UserID, &out.Text, &out.CreatedAt, &deletedAt); err != nil {
		return entity.Message{}, fmt.Errorf("DB messages.Create: %w", err)
	}
	out.DeletedAt = deletedAt
	return out, nil
}

func (r *Repo) SoftDelete(ctx context.Context, userID, messageID int64) (entity.Message, error) {
	var ownerID int64
	var deletedAt *time.Time
	if err := r.pool.QueryRow(ctx, sqlGetByID, messageID).Scan(&ownerID, &deletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Message{}, errs.NotFound{Entity: "message", ID: messageID}
		}
		return entity.Message{}, fmt.Errorf("DB messages.GetByID(id=%d): %w", messageID, err)
	}
	if deletedAt != nil {
		return entity.Message{}, errs.NotFound{Entity: "message", ID: messageID}
	}
	if ownerID != userID {
		return entity.Message{}, errs.Forbidden{Msg: "not owner"}
	}

	row := r.pool.QueryRow(ctx, sqlSoftDelete, messageID)
	var out entity.Message
	var outDeletedAt time.Time
	if err := row.Scan(&out.ID, &outDeletedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Message{}, errs.NotFound{Entity: "message", ID: messageID}
		}
		return entity.Message{}, fmt.Errorf("DB messages.SoftDelete(message_id=%d): %w", messageID, err)
	}
	out.DeletedAt = &outDeletedAt
	return out, nil
}

func (r *Repo) ListRecent(ctx context.Context, limit int) ([]entity.Message, error) {
	rows, err := r.pool.Query(ctx, sqlListRecent, limit)
	if err != nil {
		return nil, fmt.Errorf("DB messages.ListRecent(limit=%d): %w", limit, err)
	}
	defer rows.Close()

	var out []entity.Message
	for rows.Next() {
		var m entity.Message
		var deletedAt *time.Time
		if err := rows.Scan(&m.ID, &m.UserID, &m.Text, &m.CreatedAt, &deletedAt); err != nil {
			return nil, fmt.Errorf("DB messages.ListRecent: scan: %w", err)
		}
		m.DeletedAt = deletedAt
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("DB messages.ListRecent: rows: %w", err)
	}
	return out, nil
}

func (r *Repo) PurgeOld(ctx context.Context, olderThan time.Time) ([]entity.Message, error) {
	rows, err := r.pool.Query(ctx, sqlPurgeOld, olderThan)
	if err != nil {
		return nil, fmt.Errorf("DB messages.PurgeOld(older_than=%s): %w", olderThan.UTC().Format(time.RFC3339), err)
	}
	defer rows.Close()

	var out []entity.Message
	for rows.Next() {
		var m entity.Message
		var deletedAt time.Time
		if err := rows.Scan(&m.ID, &deletedAt); err != nil {
			return nil, fmt.Errorf("DB messages.PurgeOld: scan: %w", err)
		}
		m.DeletedAt = &deletedAt
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("DB messages.PurgeOld: rows: %w", err)
	}
	return out, nil
}
