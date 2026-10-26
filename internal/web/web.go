// Package web serves the site and the feed. So far: a health check, the
// private feed, the audio redirects and the voice auditions page.
package web

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"teachmepod/internal/cast"
	"teachmepod/internal/storage"
)

type Server struct {
	pool      *pgxpool.Pool
	cast      *cast.Cast
	store     *storage.Store
	feedToken string
	mux       *http.ServeMux
}

// New wires the routes. The feed and audio routes exist only when there is a
// feed token and a bucket; without them every feed URL is a 404.
func New(pool *pgxpool.Pool, c *cast.Cast, store *storage.Store, feedToken string) *Server {
	s := &Server{pool: pool, cast: c, store: store, feedToken: feedToken, mux: http.NewServeMux()}
	s.mux.HandleFunc("GET /healthz", s.healthz)
	if store != nil && feedToken != "" {
		s.mux.HandleFunc("GET /feed/{file}", s.feed)
		s.mux.HandleFunc("GET /audio/{token}/{id}/{file}", s.audio)
		s.mux.HandleFunc("GET /auditions/{token}", s.auditions)
		s.mux.HandleFunc("GET /auditions/{token}/audio/{id}", s.auditionAudio)
		s.mux.HandleFunc("POST /auditions/{token}/{id}", s.auditionVerdict)
	} else {
		slog.Warn("feed disabled: FEED_TOKEN or the audio bucket is not configured")
	}
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// healthz reports whether the database is reachable. Railway uses it to decide
// when a new deployment can take traffic.
func (s *Server) healthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.pool.Ping(ctx); err != nil {
		slog.Error("healthz: ping", "err", err)
		http.Error(w, "database unreachable", http.StatusServiceUnavailable)
		return
	}
	fmt.Fprintln(w, "ok")
}

// ListenAndServe serves on addr until ctx is cancelled, then gives in-flight
// requests up to 10 seconds to finish.
func (s *Server) ListenAndServe(ctx context.Context, addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s,
		ReadHeaderTimeout: 10 * time.Second,
	}
	errc := make(chan error, 1)
	go func() {
		slog.Info("listening", "addr", addr)
		errc <- srv.ListenAndServe()
	}()

	select {
	case err := <-errc:
		return err
	case <-ctx.Done():
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	if err := <-errc; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
