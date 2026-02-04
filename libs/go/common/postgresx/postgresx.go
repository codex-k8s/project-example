package postgresx

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Config describes Postgres connection settings.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DB       string
}

// ConfigFromEnv reads Postgres config from env without any prefix.
func ConfigFromEnv() (Config, error) { return ConfigFromEnvWithPrefix("") }

// ConfigFromEnvWithPrefix reads env vars with a prefix first and falls back to
// non-prefixed vars. This allows sharing host/port via ConfigMap while keeping
// per-service credentials in Secrets.
func ConfigFromEnvWithPrefix(prefix string) (Config, error) {
	host := strings.TrimSpace(env(prefix, "POSTGRES_HOST"))
	portStr := strings.TrimSpace(env(prefix, "POSTGRES_PORT"))
	user := strings.TrimSpace(env(prefix, "POSTGRES_USER"))
	pass := env(prefix, "POSTGRES_PASSWORD")
	db := strings.TrimSpace(env(prefix, "POSTGRES_DB"))

	if host == "" || portStr == "" || user == "" || db == "" {
		return Config{}, fmt.Errorf("postgres config: POSTGRES_HOST/POSTGRES_PORT/POSTGRES_USER/POSTGRES_DB are required")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return Config{}, fmt.Errorf("postgres config: parse POSTGRES_PORT: %w", err)
	}
	return Config{Host: host, Port: port, User: user, Password: pass, DB: db}, nil
}

func env(prefix, key string) string {
	if prefix != "" {
		if v, ok := os.LookupEnv(prefix + key); ok {
			return v
		}
	}
	return os.Getenv(key)
}

// DSN returns a Postgres DSN compatible with pgx/stdlib.
func (c Config) DSN() string {
	// Inside the cluster we typically use no TLS; extend config if required.
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", c.User, url.QueryEscape(c.Password), c.Host, c.Port, c.DB)
}

// Connect creates a pgx pool and verifies connectivity via Ping.
func Connect(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("pgxpool parse config: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("pgxpool connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pgxpool ping: %w", err)
	}
	return pool, nil
}
