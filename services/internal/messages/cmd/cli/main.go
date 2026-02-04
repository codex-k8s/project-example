package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/codex-k8s/project-example/libs/go/common/migrate"
	"github.com/codex-k8s/project-example/libs/go/common/postgresx"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cmd := strings.ToLower(os.Args[1])
	switch cmd {
	case "migrate":
		runMigrate()
	default:
		usage()
		os.Exit(2)
	}
}

func runMigrate() {
	if len(os.Args) < 3 {
		usage()
		os.Exit(2)
	}
	sub := strings.ToLower(os.Args[2])

	pgCfg, err := postgresx.ConfigFromEnvWithPrefix(strings.TrimSpace(os.Getenv("ENV_PREFIX")))
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(2)
	}

	ctx := context.Background()
	if err := migrate.RunGoose(ctx, pgCfg.DSN(), migrate.Config{
		Dir:     "cmd/cli/migrations",
		Command: sub,
		Verbose: true,
	}, os.Args[3:]...); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "Использование:")
	fmt.Fprintln(os.Stderr, "  cli migrate up")
	fmt.Fprintln(os.Stderr, "  cli migrate status")
	fmt.Fprintln(os.Stderr, "  cli migrate down")
}
