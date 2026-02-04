package app

import (
	"context"
	"fmt"
	"time"

	"github.com/codex-k8s/project-example/libs/go/common/grpcx"
	"github.com/codex-k8s/project-example/libs/go/common/logger"
	"github.com/codex-k8s/project-example/libs/go/common/otel"
	grpcgen "github.com/codex-k8s/project-example/services/jobs/messages-cleaner/internal/transport/grpc/generated/messages/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Run executes one purge operation and returns an error on failure.
func Run(ctx context.Context) error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	log := logger.New(cfg.ServiceName)

	otelShutdown, err := otel.Init(ctx, otel.ConfigFromEnv(cfg.ServiceName), log)
	if err != nil {
		log.Error("otel init failed", "err", err)
		return err
	}
	defer func() { _ = otelShutdown(context.Background()) }()

	cc, err := grpcx.Dial(ctx, cfg.MessagesGRPCAddr)
	if err != nil {
		log.Error("grpc dial failed", "err", err, "addr", cfg.MessagesGRPCAddr)
		return err
	}
	defer func() { _ = cc.Close() }()

	client := grpcgen.NewMessagesServiceClient(cc)
	olderThan := time.Now().Add(-cfg.OlderThan)

	resp, err := client.PurgeOldMessages(ctx, &grpcgen.PurgeOldMessagesRequest{
		OlderThan: timestamppb.New(olderThan),
	})
	if err != nil {
		log.Error("purge failed", "err", err, "older_than", olderThan.UTC().Format(time.RFC3339))
		return fmt.Errorf("purge old messages: %w", err)
	}

	log.Info("purge complete", "deleted", len(resp.GetDeleted()), "older_than", olderThan.UTC().Format(time.RFC3339))
	return nil
}
