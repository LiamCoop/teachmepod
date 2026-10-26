// Command podcastd runs the backlog podcast: site, feed and pipeline workers.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"

	"teachmepod/internal/auditions"
	"teachmepod/internal/cast"
	"teachmepod/internal/db"
	"teachmepod/internal/seed"
	"teachmepod/internal/storage"
	"teachmepod/internal/testrender"
	"teachmepod/internal/tts/elevenlabs"
	"teachmepod/internal/web"
)

const usage = `usage: podcastd <command>

commands:
  serve     migrate, then run the site and feed on $PORT (default 8080)
  migrate   apply pending schema migrations
  seed      load the cast and initial backlog (idempotent)
  setup     migrate, then seed
  add-test-render -profile EL-1 -title "..." file.mp3
            upload an MP3 and list it in the feed as a voice test
  auditions [-host dana] [-cast] [-model eleven_v3] [-dry-run]
            render every missing audition clip from show/hosts.yaml
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

	if err := run(ctx, os.Args[1], os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "podcastd:", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, cmd string, args []string) error {
	switch cmd {
	case "serve", "migrate", "seed", "setup", "add-test-render", "auditions":
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
		var store *storage.Store
		if os.Getenv("AUDIO_BUCKET") != "" {
			if store, err = storage.FromEnv(ctx); err != nil {
				return err
			}
		}
		c, err := cast.Load()
		if err != nil {
			return err
		}
		return web.New(pool, c, store, os.Getenv("FEED_TOKEN")).ListenAndServe(ctx, ":"+port)
	}
	if cmd == "add-test-render" {
		return addTestRender(ctx, pool, args)
	}
	if cmd == "auditions" {
		return runAuditions(ctx, pool, args)
	}
	return nil
}

func addTestRender(ctx context.Context, pool *pgxpool.Pool, args []string) error {
	fs := flag.NewFlagSet("add-test-render", flag.ContinueOnError)
	var o testrender.Options
	fs.StringVar(&o.Profile, "profile", "", "voice profile name, e.g. EL-1")
	fs.StringVar(&o.Title, "title", "", "what the clip is, e.g. \"Dana + Marcus, voice pair 1\"")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if o.Profile == "" || o.Title == "" || fs.NArg() != 1 {
		return errors.New(`usage: podcastd add-test-render -profile EL-1 -title "..." file.mp3`)
	}
	o.File = fs.Arg(0)
	store, err := storage.FromEnv(ctx)
	if err != nil {
		return err
	}
	id, err := testrender.Add(ctx, pool, store, o)
	if err != nil {
		return err
	}
	fmt.Printf("added render %d; it's in the feed now\n", id)
	return nil
}

func runAuditions(ctx context.Context, pool *pgxpool.Pool, args []string) error {
	fs := flag.NewFlagSet("auditions", flag.ContinueOnError)
	o := auditions.Options{Out: os.Stdout}
	fs.StringVar(&o.Host, "host", "", "only this host (slug); default all")
	fs.StringVar(&o.Model, "model", elevenlabs.DefaultModel, "ElevenLabs model")
	fs.BoolVar(&o.Cast, "cast", false, "only each host's cast voice")
	fs.BoolVar(&o.DryRun, "dry-run", false, "show what would be rendered, spend nothing")
	if err := fs.Parse(args); err != nil {
		return err
	}
	c, err := cast.Load()
	if err != nil {
		return err
	}
	var (
		store *storage.Store
		tts   *elevenlabs.Client
	)
	if !o.DryRun {
		if store, err = storage.FromEnv(ctx); err != nil {
			return err
		}
		if tts, err = elevenlabs.FromEnv(); err != nil {
			return err
		}
	}
	return auditions.Generate(ctx, pool, store, tts, c, o)
}
