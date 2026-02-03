package otel

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

type Config struct {
	ServiceName string
	// OTLP gRPC endpoint, например: "jaeger:4317" или "localhost:4317".
	Endpoint string
	Insecure bool
}

func ConfigFromEnv(service string) Config {
	ep := strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"))
	if ep == "" {
		ep = "jaeger:4317"
	}
	insecure := true
	switch strings.ToLower(strings.TrimSpace(os.Getenv("OTEL_EXPORTER_OTLP_INSECURE"))) {
	case "false", "0", "no":
		insecure = false
	}
	return Config{
		ServiceName: service,
		Endpoint:    ep,
		Insecure:    insecure,
	}
}

// Init инициализирует OTel tracing (OTLP gRPC exporter) и возвращает shutdown-функцию.
func Init(ctx context.Context, cfg Config, log *slog.Logger) (func(context.Context) error, error) {
	if cfg.ServiceName == "" {
		return nil, fmt.Errorf("otel: ServiceName обязателен")
	}

	opts := []otlptracegrpc.Option{otlptracegrpc.WithEndpoint(cfg.Endpoint)}
	if cfg.Insecure {
		opts = append(opts, otlptracegrpc.WithInsecure())
	}

	exp, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("otel: create exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("otel: create resource: %w", err)
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if log != nil {
		log.Info("otel: tracing enabled", "endpoint", cfg.Endpoint, "insecure", cfg.Insecure)
	}

	return tp.Shutdown, nil
}
