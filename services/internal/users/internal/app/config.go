package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Config holds runtime configuration for the users service.
type Config struct {
	ServiceName string
	GRPCPort    int
	HTTPPort    int
	EnvPrefix   string
}

// LoadConfig reads service configuration from env and validates it.
func LoadConfig() (Config, error) {
	name := strings.TrimSpace(os.Getenv("SERVICE_NAME"))
	if name == "" {
		name = "users"
	}

	envPrefix := strings.TrimSpace(os.Getenv("ENV_PREFIX"))

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
		EnvPrefix:   envPrefix,
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
