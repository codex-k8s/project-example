package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/codex-k8s/project-example/libs/go/common/grpcx"
	"github.com/codex-k8s/project-example/libs/go/common/logger"
	"github.com/codex-k8s/project-example/libs/go/common/otel"
	"github.com/codex-k8s/project-example/libs/go/common/redisx"
	"github.com/codex-k8s/project-example/libs/go/common/shutdown"
	"github.com/codex-k8s/project-example/services/external/chat-gateway/internal/domain/service"
	msggen "github.com/codex-k8s/project-example/services/external/chat-gateway/internal/transport/grpc/generated/messages/v1"
	usergen "github.com/codex-k8s/project-example/services/external/chat-gateway/internal/transport/grpc/generated/users/v1"
	httph "github.com/codex-k8s/project-example/services/external/chat-gateway/internal/transport/http"
	"github.com/codex-k8s/project-example/services/external/chat-gateway/internal/transport/http/generated"
	httpmw "github.com/codex-k8s/project-example/services/external/chat-gateway/internal/transport/http/middleware"
	"github.com/codex-k8s/project-example/services/external/chat-gateway/internal/transport/ws"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/swaggest/swgui/v5emb"
	"log/slog"
)

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

	rCfg, err := redisx.ConfigFromEnv()
	if err != nil {
		log.Error("redis config failed", "err", err)
		return err
	}
	rdb, err := redisx.Connect(ctx, rCfg)
	if err != nil {
		log.Error("redis connect failed", "err", err)
		return err
	}

	usersCC, err := grpcx.Dial(ctx, cfg.UsersGRPCAddr)
	if err != nil {
		log.Error("users grpc dial failed", "err", err, "addr", cfg.UsersGRPCAddr)
		return err
	}
	msgCC, err := grpcx.Dial(ctx, cfg.MessagesGRPCAddr)
	if err != nil {
		log.Error("messages grpc dial failed", "err", err, "addr", cfg.MessagesGRPCAddr)
		return err
	}

	users := NewUsersAdapter(usergen.NewUsersServiceClient(usersCC))
	msgs := NewMessagesAdapter(msggen.NewMessagesServiceClient(msgCC))
	sessions := NewSessionsAdapter(rdb)

	auth := service.NewAuth(users, sessions, cfg.CookieMaxAge)
	chat := service.NewChat(msgs)

	hub := ws.NewHub(log)
	wsHandler := ws.NewHandler(hub, auth, cfg.CookieName)

	// Фоновая доставка событий из messages (gRPC stream) -> WS clients.
	go forwardEvents(ctx, log, chat, hub)

	e := echo.New()
	e.HTTPErrorHandler = httpmw.ErrorHandler(log)

	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Gzip())

	validator, err := httpmw.NewOpenAPIValidator(ctx, "api/server/api.yaml")
	if err != nil {
		log.Error("openapi validator init failed", "err", err)
		return err
	}
	e.Use(validator.Middleware())

	// Tech endpoints.
	e.GET("/health/livez", func(c *echo.Context) error { return c.NoContent(http.StatusOK) })
	e.GET("/health/readyz", func(c *echo.Context) error { return c.NoContent(http.StatusOK) })
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// Swagger UI.
	e.GET("/docs/openapi.yaml", func(c *echo.Context) error { return c.File("api/server/api.yaml") })
	e.GET("/docs/*", echo.WrapHandler(v5emb.New("Chat Gateway API", "/docs/openapi.yaml", "/docs")))

	// WebSocket.
	e.GET("/ws", wsHandler.Handle)

	// OpenAPI handlers.
	api := httph.NewHandler(auth, chat, cfg.CookieName, cfg.CookieMaxAge, cfg.CookieSecure)
	generated.RegisterHandlers(e, api)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HTTPPort),
		Handler:           e,
		ReadHeaderTimeout: 5 * time.Second,
	}

	serveErr := make(chan error, 1)
	go func() {
		log.Info("http server started", "port", cfg.HTTPPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serveErr <- err
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
	err = shutdown.Run(shutdownCtx,
		func(ctx context.Context) error { return srv.Shutdown(ctx) },
		func(context.Context) error { _ = usersCC.Close(); return nil },
		func(context.Context) error { _ = msgCC.Close(); return nil },
		func(context.Context) error { _ = rdb.Close(); return nil },
		otelShutdown,
	)
	if err != nil {
		logger.WithContext(ctx, log).Error("shutdown finished with error", "err", err)
		return err
	}
	log.Info("shutdown complete")
	return nil
}

func forwardEvents(ctx context.Context, log *slog.Logger, chat *service.Chat, hub *ws.Hub) {
	backoff := 1 * time.Second
	for {
		if ctx.Err() != nil {
			return
		}
		ch, err := chat.Subscribe(ctx)
		if err != nil {
			logger.WithContext(ctx, log).Error("subscribe failed", "err", err)
			time.Sleep(backoff)
			continue
		}
		for evt := range ch {
			b, err := ws.EncodeEvent(evt)
			if err != nil {
				logger.WithContext(ctx, log).Warn("encode ws event failed", "err", err)
				continue
			}
			hub.Broadcast(ctx, b)
		}
		// stream завершился — попробуем переподключиться.
		time.Sleep(backoff)
	}
}
