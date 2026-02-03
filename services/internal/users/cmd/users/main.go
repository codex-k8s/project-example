package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/codex-k8s/project-example/services/internal/users/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(ctx); err != nil {
		// Логирование делается внутри app (граница); тут только exit-code.
		os.Exit(1)
	}
}
