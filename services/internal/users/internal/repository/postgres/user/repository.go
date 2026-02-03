package user

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/codex-k8s/project-example/services/internal/users/internal/domain/errs"
	userrepo "github.com/codex-k8s/project-example/services/internal/users/internal/domain/repository/user"
	"github.com/codex-k8s/project-example/services/internal/users/internal/domain/types/entity"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed sql/create_user.sql
var sqlCreateUser string

//go:embed sql/get_user_by_username.sql
var sqlGetByUsername string

//go:embed sql/get_user_by_id.sql
var sqlGetByID string

type Repo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repo {
	return &Repo{pool: pool}
}

var _ userrepo.Repository = (*Repo)(nil)

func (r *Repo) Create(ctx context.Context, u entity.User) (entity.User, error) {
	row := r.pool.QueryRow(ctx, sqlCreateUser, u.Username, u.PasswordHash)
	var out entity.User
	var createdAt time.Time
	if err := row.Scan(&out.ID, &out.Username, &createdAt); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return entity.User{}, errs.Conflict{Msg: "username already exists"}
		}
		return entity.User{}, fmt.Errorf("DB users.Create: %w", err)
	}
	out.CreatedAt = createdAt
	out.PasswordHash = u.PasswordHash
	return out, nil
}

func (r *Repo) GetByUsername(ctx context.Context, username string) (entity.User, error) {
	return r.getOne(ctx, sqlGetByUsername, []any{username}, username, "GetByUsername")
}

func (r *Repo) GetByID(ctx context.Context, id int64) (entity.User, error) {
	return r.getOne(ctx, sqlGetByID, []any{id}, id, "GetByID")
}

func (r *Repo) getOne(ctx context.Context, query string, args []any, notFoundID any, op string) (entity.User, error) {
	row := r.pool.QueryRow(ctx, query, args...)
	var out entity.User
	var createdAt time.Time
	if err := row.Scan(&out.ID, &out.Username, &out.PasswordHash, &createdAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return entity.User{}, errs.NotFound{Entity: "user", ID: notFoundID}
		}
		return entity.User{}, fmt.Errorf("DB users.%s: %w", op, err)
	}
	out.CreatedAt = createdAt
	return out, nil
}
