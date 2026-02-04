package middleware

import (
	"context"
	"errors"
	"log/slog"

	"github.com/codex-k8s/project-example/libs/go/common/logger"
	"github.com/codex-k8s/project-example/services/internal/messages/internal/domain/errs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// UnaryErrorBoundary maps domain errors to gRPC status codes and logs once (unary RPC).
func UnaryErrorBoundary(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err == nil {
			return resp, nil
		}
		return nil, mapError(ctx, log, info.FullMethod, err)
	}
}

// StreamErrorBoundary maps domain errors to gRPC status codes and logs once (streaming RPC).
func StreamErrorBoundary(log *slog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		err := handler(srv, ss)
		if err == nil {
			return nil
		}
		return mapError(ss.Context(), log, info.FullMethod, err)
	}
}

func mapError(ctx context.Context, log *slog.Logger, method string, err error) error {
	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, "canceled")
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	}

	var v errs.Validation
	if errors.As(err, &v) {
		return status.Error(codes.InvalidArgument, "validation error")
	}
	var nf errs.NotFound
	if errors.As(err, &nf) {
		return status.Error(codes.NotFound, "not found")
	}
	var f errs.Forbidden
	if errors.As(err, &f) {
		return status.Error(codes.PermissionDenied, "forbidden")
	}

	logger.WithContext(ctx, log).Error("grpc request failed", "method", method, "err", err)
	return status.Error(codes.Internal, "internal error")
}
