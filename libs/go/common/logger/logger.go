package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel/trace"
)

// New creates a JSON slog logger and adds a "service" field.
// Log level is controlled via LOG_LEVEL=debug|info|warn|error.
func New(service string) *slog.Logger {
	level := slog.LevelInfo
	switch strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL"))) {
	case "debug":
		level = slog.LevelDebug
	case "info", "":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	h := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	log := slog.New(h)
	if service != "" {
		log = log.With("service", service)
	}
	return log
}

// WithContext enriches log with trace/span IDs from the current OTEL span context (if present).
func WithContext(ctx context.Context, log *slog.Logger) *slog.Logger {
	if ctx == nil {
		return log
	}
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return log
	}
	return log.With(
		"trace_id", strings.ToLower(sc.TraceID().String()),
		"span_id", strings.ToLower(sc.SpanID().String()),
	)
}
