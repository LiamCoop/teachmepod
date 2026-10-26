package web

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"teachmepod/internal/feed"
)

// Presigned audio links last long enough for a slow download or a resume
// mid-walk. An app holding an expired one asks the stable URL again.
const audioLinkTTL = 6 * time.Hour

// listed is the condition for a render to appear in the feed: finished, and
// either a voice test or the render its episode published.
const listed = `r.status = 'done' AND r.audio_key IS NOT NULL AND r.audio_bytes IS NOT NULL
	AND (r.is_test OR e.published_render_id = r.id)`

// validToken compares in constant time so the token can't be guessed byte by byte.
func (s *Server) validToken(got string) bool {
	return s.feedToken != "" && subtle.ConstantTimeCompare([]byte(got), []byte(s.feedToken)) == 1
}

// feed serves /feed/<token>.xml.
func (s *Server) feed(w http.ResponseWriter, r *http.Request) {
	if !s.validToken(tokenFromFile(r.PathValue("file"))) {
		http.NotFound(w, r)
		return
	}
	base := baseURL(r)
	rows, err := s.pool.Query(r.Context(), `
		SELECT r.id, r.audio_bytes, r.duration_s, r.is_test, vp.name,
		       COALESCE(NULLIF(e.title, ''), t.title), COALESCE(e.slug, ''),
		       COALESCE(e.published_at, r.finished_at, r.created_at) AS at
		FROM renders r
		JOIN episodes e ON e.id = r.episode_id
		JOIN topics t ON t.id = e.topic_id
		JOIN voice_profiles vp ON vp.id = r.voice_profile_id
		WHERE `+listed+`
		ORDER BY at DESC
		LIMIT 200`)
	if err != nil {
		serverError(w, "feed: query", err)
		return
	}
	ch := feed.Channel{
		Title:       "teachmepod",
		Description: "A private two-host technical podcast, made from a backlog.",
		Link:        base,
	}
	for rows.Next() {
		var (
			id, bytes   int64
			durationS   *int32
			isTest      bool
			profile     string
			title, slug string
			at          time.Time
		)
		if err := rows.Scan(&id, &bytes, &durationS, &isTest, &profile, &title, &slug, &at); err != nil {
			serverError(w, "feed: scan", err)
			return
		}
		switch {
		case isTest:
			slug = "test-" + strings.ToLower(profile)
		case slug == "":
			slug = "episode"
		}
		it := feed.Item{
			GUID:        fmt.Sprintf("teachmepod-render-%d", id),
			Title:       title,
			Description: title,
			AudioURL:    fmt.Sprintf("%s/audio/%s/%d/%s.mp3", base, s.feedToken, id, slug),
			AudioBytes:  bytes,
			Published:   at,
		}
		if isTest {
			it.Title = fmt.Sprintf("Test: %s · %s", profile, title)
			it.Description = fmt.Sprintf("Voice test, profile %s.", profile)
		}
		if durationS != nil {
			it.Duration = time.Duration(*durationS) * time.Second
		}
		ch.Items = append(ch.Items, it)
	}
	if err := rows.Err(); err != nil {
		serverError(w, "feed: rows", err)
		return
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.Header().Set("Cache-Control", "private, no-cache")
	if err := feed.Write(w, ch); err != nil {
		slog.Error("feed: write", "err", err)
	}
}

// audio serves /audio/<token>/<render id>/<slug>.mp3 by redirecting to a
// presigned bucket URL, so the bytes come straight from the bucket. The slug
// is only there to give the file a readable name.
func (s *Server) audio(w http.ResponseWriter, r *http.Request) {
	if !s.validToken(r.PathValue("token")) {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	var key string
	err = s.pool.QueryRow(r.Context(), `
		SELECT r.audio_key FROM renders r JOIN episodes e ON e.id = r.episode_id
		WHERE r.id = $1 AND `+listed, id).Scan(&key)
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		serverError(w, "audio: query", err)
		return
	}
	u, err := s.store.PresignGet(r.Context(), key, audioLinkTTL)
	if err != nil {
		serverError(w, "audio: presign", err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, u, http.StatusFound)
}

// baseURL is the site's own origin as the listener reached it. Railway
// terminates TLS and passes the original scheme in X-Forwarded-Proto.
func baseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// tokenFromFile turns "<token>.xml" into the token, or "" for anything else.
func tokenFromFile(file string) string {
	token, ok := strings.CutSuffix(file, ".xml")
	if !ok {
		return ""
	}
	return token
}

func serverError(w http.ResponseWriter, msg string, err error) {
	slog.Error(msg, "err", err)
	http.Error(w, "internal error", http.StatusInternalServerError)
}
