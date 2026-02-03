package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type Config struct {
	Driver       string // sql driver name (по умолчанию: "pgx")
	Dir          string // directory with *.sql
	Command      string // "up", "down", "status", ...
	Verbose      bool
	AllowMissing bool
}

// RunGoose запускает goose как библиотеку (без внешнего бинарника).
// DSN должен быть полноценным (postgres://...).
func RunGoose(ctx context.Context, dsn string, cfg Config, args ...string) error {
	if cfg.Driver == "" {
		cfg.Driver = "pgx"
	}
	if cfg.Dir == "" {
		return fmt.Errorf("goose: Dir обязателен")
	}
	if cfg.Command == "" {
		return fmt.Errorf("goose: Command обязателен")
	}

	goose.SetVerbose(cfg.Verbose)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose: set dialect: %w", err)
	}
	goose.SetBaseFS(os.DirFS(cfg.Dir))
	// goose expects a directory path relative to baseFS.
	dir := "."

	db, err := sql.Open(cfg.Driver, dsn)
	if err != nil {
		return fmt.Errorf("goose: open db: %w", err)
	}
	defer func() { _ = db.Close() }()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("goose: ping db: %w", err)
	}

	// Поддерживаем минимум команд, остальное - через goose.Run.
	if strings.EqualFold(cfg.Command, "up") {
		if err := goose.Up(db, dir); err != nil {
			if cfg.AllowMissing && strings.Contains(strings.ToLower(err.Error()), "no migrations") {
				return nil
			}
			return fmt.Errorf("goose up: %w", err)
		}
		return nil
	}

	if err := goose.RunContext(ctx, cfg.Command, db, dir, args...); err != nil {
		return fmt.Errorf("goose %s: %w", cfg.Command, err)
	}
	return nil
}
