package middleware

import (
	"context"
	"errors"
	"log/slog"

	"github.com/codex-k8s/project-example/libs/go/common/logger"
	"github.com/codex-k8s/project-example/services/internal/users/internal/domain/errs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryErrorBoundary is the only boundary: it maps domain errors to gRPC status codes and logs once.
func UnaryErrorBoundary(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}

		// Context errors are not logged as ERROR.
		if errors.Is(err, context.Canceled) {
			return nil, status.Error(codes.Canceled, "canceled")
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, status.Error(codes.DeadlineExceeded, "deadline exceeded")
		}

		// Domain -> gRPC codes.
		var v errs.Validation
		if errors.As(err, &v) {
			return nil, status.Error(codes.InvalidArgument, "validation error")
		}
		var u errs.Unauthorized
		if errors.As(err, &u) {
			return nil, status.Error(codes.Unauthenticated, "unauthorized")
		}
		var nf errs.NotFound
		if errors.As(err, &nf) {
			return nil, status.Error(codes.NotFound, "not found")
		}
		var c errs.Conflict
		if errors.As(err, &c) {
			return nil, status.Error(codes.AlreadyExists, "conflict")
		}

		// Everything else is Internal.
		logger.WithContext(ctx, log).Error("grpc request failed", "method", info.FullMethod, "err", err)
		return nil, status.Error(codes.Internal, "internal error")
	}
}
