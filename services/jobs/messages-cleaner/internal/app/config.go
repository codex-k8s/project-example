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
	// Target gRPC address of messages service, e.g. "messages:8080".
	MessagesGRPCAddr string
	OlderThan        time.Duration
}

func LoadConfig() (Config, error) {
	name := strings.TrimSpace(os.Getenv("SERVICE_NAME"))
	if name == "" {
		name = "messages-cleaner"
	}

	addr := strings.TrimSpace(os.Getenv("MESSAGES_GRPC_ADDR"))
	if addr == "" {
		addr = "messages:8080"
	}

	secs, err := mustInt("OLDER_THAN_SECONDS", 600)
	if err != nil {
		return Config{}, err
	}

	return Config{
		ServiceName:      name,
		MessagesGRPCAddr: addr,
		OlderThan:        time.Duration(secs) * time.Second,
	}, nil
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
