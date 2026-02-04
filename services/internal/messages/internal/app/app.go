package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/codex-k8s/project-example/libs/go/common/grpcx"
	"github.com/codex-k8s/project-example/libs/go/common/logger"
	"github.com/codex-k8s/project-example/libs/go/common/otel"
	"github.com/codex-k8s/project-example/libs/go/common/postgresx"
	"github.com/codex-k8s/project-example/libs/go/common/shutdown"
	msgsvc "github.com/codex-k8s/project-example/services/internal/messages/internal/domain/service"
	msgrepo "github.com/codex-k8s/project-example/services/internal/messages/internal/repository/postgres/message"
	grpcsvc "github.com/codex-k8s/project-example/services/internal/messages/internal/transport/grpc"
	grpcmw "github.com/codex-k8s/project-example/services/internal/messages/internal/transport/grpc/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func Run(ctx context.Context) (runErr error) {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	log := logger.New(cfg.ServiceName)

	var closers []shutdown.Closer
	didShutdown := false
	defer func() {
		if runErr == nil || didShutdown {
			return
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		_ = shutdown.Run(shutdownCtx, closers...)
	}()

	otelShutdown, err := otel.Init(ctx, otel.ConfigFromEnv(cfg.ServiceName), log)
	if err != nil {
		log.Error("otel init failed", "err", err)
		return err
	}
	closers = append(closers, otelShutdown)

	pgCfg, err := postgresx.ConfigFromEnvWithPrefix(cfg.EnvPrefix)
	if err != nil {
		log.Error("postgres config failed", "err", err)
		return err
	}
	pool, err := postgresx.Connect(ctx, pgCfg)
	if err != nil {
		log.Error("postgres connect failed", "err", err)
		return err
	}
	closers = append(closers, func(context.Context) error { pool.Close(); return nil })

	repo := msgrepo.New(pool)
	svc := msgsvc.New(repo)

	grpcLis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		log.Error("grpc listen failed", "err", err)
		return err
	}
	closers = append(closers, func(context.Context) error { _ = grpcLis.Close(); return nil })

	grpcServer := grpcx.NewServer(
		log,
		[]grpc.UnaryServerInterceptor{grpcmw.UnaryErrorBoundary(log)},
		[]grpc.StreamServerInterceptor{grpcmw.StreamErrorBoundary(log)},
	)
	grpcsvc.Register(grpcServer, svc)
	closers = append(closers, func(ctx context.Context) error {
		stopped := make(chan struct{})
		go func() {
			grpcServer.GracefulStop()
			close(stopped)
		}()
		select {
		case <-ctx.Done():
			grpcServer.Stop()
			return ctx.Err()
		case <-stopped:
			return nil
		}
	})

	httpSrv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		ReadHeaderTimeout: 5 * time.Second,
		Handler:           http.NewServeMux(),
	}
	closers = append(closers, func(ctx context.Context) error { return httpSrv.Shutdown(ctx) })
	mux := httpSrv.Handler.(*http.ServeMux)
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health/livez", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	mux.HandleFunc("/health/readyz", func(w http.ResponseWriter, r *http.Request) {
		c, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()
		if err := pool.Ping(c); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	serveErr := make(chan error, 2)
	go func() {
		log.Info("grpc server started", "port", cfg.GRPCPort)
		if err := grpcServer.Serve(grpcLis); err != nil {
			serveErr <- fmt.Errorf("grpc serve: %w", err)
		}
	}()
	go func() {
		log.Info("http server started", "port", cfg.HTTPPort)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serveErr <- fmt.Errorf("http serve: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-serveErr:
		log.Error("server failed", "err", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	log.Info("shutting down")
	didShutdown = true
	err = shutdown.Run(shutdownCtx, closers...)
	if err != nil {
		logger.WithContext(ctx, log).Error("shutdown finished with error", "err", err)
		return err
	}
	log.Info("shutdown complete")
	return nil
}
