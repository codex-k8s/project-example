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

type ownerRow struct {
	UserID    int64      `db:"user_id"`
	DeletedAt *time.Time `db:"deleted_at"`
}

type deleteRow struct {
	ID        int64     `db:"id"`
	DeletedAt time.Time `db:"deleted_at"`
}

func (r *Repo) Create(ctx context.Context, msg entity.Message) (entity.Message, error) {
	row := r.pool.QueryRow(ctx, sqlCreate, msg.UserID, msg.Text)
	cr, ok := row.(pgx.CollectableRow)
	if !ok {
		return entity.Message{}, fmt.Errorf("DB messages.Create: unexpected row type")
	}
	out, err := pgx.RowToStructByName[entity.Message](cr)
	if err != nil {
		return entity.Message{}, fmt.Errorf("DB messages.Create: %w", err)
	}
	return out, nil
}

func (r *Repo) SoftDelete(ctx context.Context, userID, messageID int64) (entity.Message, error) {
	row := r.pool.QueryRow(ctx, sqlGetByID, messageID)
	cr, ok := row.(pgx.CollectableRow)
	if !ok {
		return entity.Message{}, fmt.Errorf("DB messages.GetByID: unexpected row type")
	}
	own, err := pgx.RowToStructByName[ownerRow](cr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Message{}, errs.NotFound{Entity: "message", ID: messageID}
		}
		return entity.Message{}, fmt.Errorf("DB messages.GetByID: %w", err)
	}
	if own.DeletedAt != nil {
		return entity.Message{}, errs.NotFound{Entity: "message", ID: messageID}
	}
	if own.UserID != userID {
		return entity.Message{}, errs.Forbidden{Msg: "not owner"}
	}

	row = r.pool.QueryRow(ctx, sqlSoftDelete, messageID)
	cr, ok = row.(pgx.CollectableRow)
	if !ok {
		return entity.Message{}, fmt.Errorf("DB messages.SoftDelete: unexpected row type")
	}
	del, err := pgx.RowToStructByName[deleteRow](cr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.Message{}, errs.NotFound{Entity: "message", ID: messageID}
		}
		return entity.Message{}, fmt.Errorf("DB messages.SoftDelete: %w", err)
	}
	out := entity.Message{ID: del.ID, DeletedAt: &del.DeletedAt}
	return out, nil
}

func (r *Repo) ListRecent(ctx context.Context, limit int) ([]entity.Message, error) {
	rows, err := r.pool.Query(ctx, sqlListRecent, limit)
	if err != nil {
		return nil, fmt.Errorf("DB messages.ListRecent(limit=%d): %w", limit, err)
	}
	defer rows.Close()

	out, err := pgx.CollectRows(rows, pgx.RowToStructByName[entity.Message])
	if err != nil {
		return nil, fmt.Errorf("DB messages.ListRecent: %w", err)
	}
	return out, nil
}

func (r *Repo) PurgeOld(ctx context.Context, olderThan time.Time) ([]entity.Message, error) {
	rows, err := r.pool.Query(ctx, sqlPurgeOld, olderThan)
	if err != nil {
		return nil, fmt.Errorf("DB messages.PurgeOld(older_than=%s): %w", olderThan.UTC().Format(time.RFC3339), err)
	}
	defer rows.Close()

	delRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[deleteRow])
	if err != nil {
		return nil, fmt.Errorf("DB messages.PurgeOld: %w", err)
	}
	out := make([]entity.Message, 0, len(delRows))
	for _, d := range delRows {
		dt := d.DeletedAt
		out = append(out, entity.Message{ID: d.ID, DeletedAt: &dt})
	}
	return out, nil
}
