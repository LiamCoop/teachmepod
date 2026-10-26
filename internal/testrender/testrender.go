// Package testrender adds a hand-made MP3 to the feed as a voice test, so a
// voice can be heard on a real walk before any pipeline exists to render it.
// Each one hangs off a parked "Voice tests" topic and its own episode.
package testrender

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"teachmepod/internal/audio"
	"teachmepod/internal/storage"
)

type Options struct {
	File    string // path to an MP3
	Profile string // voice profile name, e.g. EL-1
	Title   string
}

// Add uploads the file and lists it in the feed. It returns the render ID.
func Add(ctx context.Context, pool *pgxpool.Pool, store *storage.Store, o Options) (int64, error) {
	if !strings.EqualFold(filepath.Ext(o.File), ".mp3") {
		return 0, fmt.Errorf("%s: only MP3 files are supported", o.File)
	}
	var profileID int64
	err := pool.QueryRow(ctx, `SELECT id FROM voice_profiles WHERE name = $1`, o.Profile).Scan(&profileID)
	if err != nil {
		return 0, fmt.Errorf("voice profile %q: %w", o.Profile, err)
	}

	f, err := os.Open(o.File)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return 0, err
	}
	duration, err := audio.ProbeSeconds(ctx, o.File)
	if err != nil {
		return 0, err
	}

	key := fmt.Sprintf("test/%s/%s-%s", o.Profile, time.Now().UTC().Format("20060102T150405Z"), filepath.Base(o.File))
	if err := store.Put(ctx, key, "audio/mpeg", f, info.Size()); err != nil {
		return 0, err
	}

	var renderID int64
	err = pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		var topicID int64
		err := tx.QueryRow(ctx, `
			INSERT INTO topics (title, status, baseline)
			VALUES ('Voice tests', 'parked', 'Holds hand-made voice test renders; never picked.')
			ON CONFLICT (lower(title)) DO UPDATE SET updated_at = topics.updated_at
			RETURNING id`).Scan(&topicID)
		if err != nil {
			return fmt.Errorf("voice tests topic: %w", err)
		}
		var episodeID int64
		err = tx.QueryRow(ctx, `
			INSERT INTO episodes (topic_id, status, status_reason, title)
			VALUES ($1, 'parked', 'voice test', $2) RETURNING id`,
			topicID, o.Title).Scan(&episodeID)
		if err != nil {
			return fmt.Errorf("episode: %w", err)
		}
		return tx.QueryRow(ctx, `
			INSERT INTO renders (episode_id, voice_profile_id, status, audio_key, audio_bytes,
			                     duration_s, is_test, finished_at)
			VALUES ($1, $2, 'done', $3, $4, $5, true, now()) RETURNING id`,
			episodeID, profileID, key, info.Size(), duration).Scan(&renderID)
	})
	if err != nil {
		return 0, fmt.Errorf("record render (uploaded as %s): %w", key, err)
	}
	return renderID, nil
}
