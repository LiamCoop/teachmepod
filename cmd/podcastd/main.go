// Command podcastd runs the backlog podcast: site, feed and pipeline workers.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"teachmepod/internal/db"
	"teachmepod/internal/seed"
)

const usage = `usage: podcastd <command>

commands:
  migrate   apply pending schema migrations
  seed      load the cast and initial backlog (idempotent)
  setup     migrate, then seed
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	if err := run(ctx, os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "podcastd:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cmd string) error {
	switch cmd {
	case "migrate", "seed", "setup":
	default:
		fmt.Fprint(os.Stderr, usage)
		return fmt.Errorf("unknown command %q", cmd)
	}

	pool, err := db.Open(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	if cmd == "migrate" || cmd == "setup" {
		if err := db.Migrate(ctx, pool); err != nil {
			return err
		}
	}
	if cmd == "seed" || cmd == "setup" {
		if err := seed.Run(ctx, pool); err != nil {
			return err
		}
	}
	return nil
}
