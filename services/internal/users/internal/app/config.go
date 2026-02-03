package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServiceName string
	GRPCPort    int
	HTTPPort    int
}

func LoadConfig() (Config, error) {
	name := strings.TrimSpace(os.Getenv("SERVICE_NAME"))
	if name == "" {
		name = "users"
	}

	grpcPort, err := mustPort("GRPC_PORT", 8080)
	if err != nil {
		return Config{}, err
	}
	httpPort, err := mustPort("HTTP_PORT", 8081)
	if err != nil {
		return Config{}, err
	}

	return Config{
		ServiceName: name,
		GRPCPort:    grpcPort,
		HTTPPort:    httpPort,
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
