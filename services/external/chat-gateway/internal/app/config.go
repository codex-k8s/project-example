package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ServiceName string
	HTTPPort    int

	UsersGRPCAddr    string
	MessagesGRPCAddr string

	CookieName   string
	CookieMaxAge time.Duration
	CookieSecure bool

	EnvPrefix string
}

func LoadConfig() (Config, error) {
	name := strings.TrimSpace(os.Getenv("SERVICE_NAME"))
	if name == "" {
		name = "chat-gateway"
	}

	envPrefix := strings.TrimSpace(os.Getenv("ENV_PREFIX"))

	httpPort, err := mustPort("HTTP_PORT", 8080)
	if err != nil {
		return Config{}, err
	}

	usersAddr := strings.TrimSpace(os.Getenv("USERS_GRPC_ADDR"))
	if usersAddr == "" {
		usersAddr = "users:8080"
	}
	msgAddr := strings.TrimSpace(os.Getenv("MESSAGES_GRPC_ADDR"))
	if msgAddr == "" {
		msgAddr = "messages:8080"
	}

	cookieName := strings.TrimSpace(os.Getenv("COOKIE_NAME"))
	if cookieName == "" {
		cookieName = "sid"
	}

	maxAgeSec, err := mustInt("COOKIE_MAX_AGE_SECONDS", 86400)
	if err != nil {
		return Config{}, err
	}

	secure := true
	switch strings.ToLower(strings.TrimSpace(os.Getenv("COOKIE_SECURE"))) {
	case "false", "0", "no":
		secure = false
	case "":
		// В dev можно оставить не-secure, но по умолчанию лучше безопасно.
	}

	return Config{
		ServiceName:      name,
		HTTPPort:         httpPort,
		UsersGRPCAddr:    usersAddr,
		MessagesGRPCAddr: msgAddr,
		CookieName:       cookieName,
		CookieMaxAge:     time.Duration(maxAgeSec) * time.Second,
		CookieSecure:     secure,
		EnvPrefix:        envPrefix,
	}, nil
}

func mustPort(env string, def int) (int, error) {
	v := strings.TrimSpace(os.Getenv(env))
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s: parse int: %w", env, err)
	}
	if n <= 0 || n > 65535 {
		return 0, fmt.Errorf("%s: invalid port: %d", env, n)
	}
	return n, nil
}

func mustInt(env string, def int) (int, error) {
	v := strings.TrimSpace(os.Getenv(env))
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("%s: parse int: %w", env, err)
	}
	return n, nil
}
