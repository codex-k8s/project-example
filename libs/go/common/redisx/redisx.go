package redisx

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// ConfigFromEnvWithPrefix reads env vars with a prefix first and falls back to
// non-prefixed vars. Use it to keep per-service secrets isolated.
func ConfigFromEnvWithPrefix(prefix string) (Config, error) {
	host := strings.TrimSpace(env(prefix, "REDIS_HOST"))
	portStr := strings.TrimSpace(env(prefix, "REDIS_PORT"))
	pass := env(prefix, "REDIS_PASSWORD")
	dbStr := strings.TrimSpace(env(prefix, "REDIS_DB"))

	if host == "" || portStr == "" {
		return Config{}, fmt.Errorf("redis config: REDIS_HOST/REDIS_PORT обязательны")
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return Config{}, fmt.Errorf("redis config: parse REDIS_PORT: %w", err)
	}
	db := 0
	if dbStr != "" {
		db, err = strconv.Atoi(dbStr)
		if err != nil {
			return Config{}, fmt.Errorf("redis config: parse REDIS_DB: %w", err)
		}
	}
	return Config{Host: host, Port: port, Password: pass, DB: db}, nil
}

func env(prefix, key string) string {
	if prefix != "" {
		if v, ok := os.LookupEnv(prefix + key); ok {
			return v
		}
	}
	return os.Getenv(key)
}

func Connect(ctx context.Context, cfg Config) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return rdb, nil
}
