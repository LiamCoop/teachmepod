// Command podcastd runs the backlog podcast: site, feed and pipeline workers.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"teachmepod/internal/db"
	"teachmepod/internal/seed"
	"teachmepod/internal/web"
)

const usage = `usage: podcastd <command>

commands:
  serve     migrate, then run the site and feed on $PORT (default 8080)
  migrate   apply pending schema migrations
  seed      load the cast and initial backlog (idempotent)
  setup     migrate, then seed
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	// Railway marks anything on stderr as an error, so routine logs go to stdout.
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, nil)))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := run(ctx, os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "podcastd:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cmd string) error {
	switch cmd {
	case "serve", "migrate", "seed", "setup":
	default:
		fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("unknown command %q", cmd)
	}

	pool, err := db.Open(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	if cmd == "serve" || cmd == "migrate" || cmd == "setup" {
		if err := db.Migrate(ctx, pool); err != nil {
			return err
		}
	}
	if cmd == "seed" || cmd == "setup" {
		if err := seed.Run(ctx, pool); err != nil {
			return err
		}
	}
	if cmd == "serve" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		return web.New(pool).ListenAndServe(ctx, ":"+port)
	}
	return nil
}
