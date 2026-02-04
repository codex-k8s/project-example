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

// Config defines goose execution settings (when used as a library).
type Config struct {
	Driver       string // SQL driver name (default: "pgx").
	Dir          string // Directory containing migration *.sql files.
	Command      string // goose command: "up", "down", "status", ...
	Verbose      bool
	AllowMissing bool
}

// RunGoose runs goose as a library (no external goose binary required).
// dsn must be a full Postgres DSN (e.g. postgres://...).
func RunGoose(ctx context.Context, dsn string, cfg Config, args ...string) error {
	if cfg.Driver == "" {
		cfg.Driver = "pgx"
	}
	if cfg.Dir == "" {
		return fmt.Errorf("goose: Dir is required")
	}
	if cfg.Command == "" {
		return fmt.Errorf("goose: Command is required")
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

	// Support "up" explicitly to allow graceful handling of "no migrations" when configured.
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
