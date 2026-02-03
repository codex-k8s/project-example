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

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DB       string
}

func ConfigFromEnv() (Config, error) {
	host := strings.TrimSpace(os.Getenv("POSTGRES_HOST"))
	portStr := strings.TrimSpace(os.Getenv("POSTGRES_PORT"))
	user := strings.TrimSpace(os.Getenv("POSTGRES_USER"))
	pass := os.Getenv("POSTGRES_PASSWORD")
	db := strings.TrimSpace(os.Getenv("POSTGRES_DB"))

	if host == "" || portStr == "" || user == "" || db == "" {
		return Config{}, fmt.Errorf("postgres config: POSTGRES_HOST/POSTGRES_PORT/POSTGRES_USER/POSTGRES_DB обязательны")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return Config{}, fmt.Errorf("postgres config: parse POSTGRES_PORT: %w", err)
	}
	return Config{Host: host, Port: port, User: user, Password: pass, DB: db}, nil
}

func (c Config) DSN() string {
	// Внутри кластера обычно используем без TLS; если нужно иначе — расширим конфиг.
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", c.User, url.QueryEscape(c.Password), c.Host, c.Port, c.DB)
}

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
