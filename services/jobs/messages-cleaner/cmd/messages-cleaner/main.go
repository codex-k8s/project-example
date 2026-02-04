package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/codex-k8s/project-example/services/jobs/messages-cleaner/internal/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	defer func() {
		signal.Stop(sigCh)
		cancel()
	}()

	go func() {
		<-sigCh
		cancel()
		<-sigCh
		os.Exit(2)
	}()

	if err := app.Run(ctx); err != nil {
		os.Exit(1)
	}
}
