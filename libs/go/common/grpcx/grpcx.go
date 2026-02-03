package grpcx

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// NewServer создаёт gRPC server с базовыми interceptor'ами: recovery + OTEL.
// Дополнительные interceptor'ы (например, error mapping) передаются отдельно и будут
// выполнены *внутри* OTEL (чтобы OTEL увидел уже замаппленную ошибку/status).
func NewServer(
	log *slog.Logger,
	unaryExtra []grpc.UnaryServerInterceptor,
	streamExtra []grpc.StreamServerInterceptor,
	extra ...grpc.ServerOption,
) *grpc.Server {
	unaryChain := []grpc.UnaryServerInterceptor{
		unaryRecovery(log),
	}
	unaryChain = append(unaryChain, unaryExtra...)

	streamChain := []grpc.StreamServerInterceptor{
		streamRecovery(log),
	}
	streamChain = append(streamChain, streamExtra...)

	opts := []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(unaryChain...),
		grpc.ChainStreamInterceptor(streamChain...),
	}
	opts = append(opts, extra...)
	return grpc.NewServer(opts...)
}

func Dial(ctx context.Context, target string, extra ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
	opts = append(opts, extra...)
	cc, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("grpc dial %q: %w", target, err)
	}

	// Явно дёргаем Connect + ждём готовности, чтобы ошибки "плохого target"
	// проявлялись в месте инициализации, а не внутри первого RPC.
	cc.Connect()
	for {
		st := cc.GetState()
		if st == connectivity.Ready {
			break
		}
		if !cc.WaitForStateChange(ctx, st) {
			_ = cc.Close()
			return nil, fmt.Errorf("grpc dial %q: %w", target, ctx.Err())
		}
	}
	return cc, nil
}

func unaryRecovery(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var (
			resp any
			err  error
		)
		defer func() {
			if r := recover(); r != nil {
				if log != nil {
					log.Error("panic recovered (grpc unary)", "method", info.FullMethod, "panic", r, "stack", string(debug.Stack()))
				}
				resp = nil
				err = status.Error(codes.Internal, "internal error")
			}
		}()
		resp, err = handler(ctx, req)
		return resp, err
	}
}

func streamRecovery(log *slog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var err error
		defer func() {
			if r := recover(); r != nil {
				if log != nil {
					log.Error("panic recovered (grpc stream)", "method", info.FullMethod, "panic", r, "stack", string(debug.Stack()))
				}
				err = status.Error(codes.Internal, "internal error")
			}
		}()
		err = handler(srv, ss)
		return err
	}
}
