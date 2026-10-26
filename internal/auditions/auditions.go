// Package auditions renders each host's audition line in every shortlisted
// voice from show/hosts.yaml and stores the clips for review. A clip that
// already exists for the same host, voice, model and line is never paid for twice.
package auditions

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/jackc/pgx/v5/pgxpool"

	"teachmepod/internal/audio"
	"teachmepod/internal/cast"
	"teachmepod/internal/storage"
	"teachmepod/internal/tts/elevenlabs"
)

type Options struct {
	Host   string // only this host; empty for all
	Model  string
	DryRun bool // list what would be rendered and its size, spend nothing
	Out    io.Writer
}

type job struct {
	host  cast.Host
	voice cast.VoiceOption
	sha   string
}

// Generate renders every missing audition. tts may be nil when DryRun is set.
func Generate(ctx context.Context, pool *pgxpool.Pool, store *storage.Store, tts *elevenlabs.Client, c *cast.Cast, o Options) error {
	if o.Host != "" {
		if _, ok := c.Host(o.Host); !ok {
			return fmt.Errorf("unknown host %q", o.Host)
		}
	}

	var todo []job
	var skipped, chars int
	for _, h := range c.Hosts {
		if o.Host != "" && h.Slug != o.Host {
			continue
		}
		sum := sha256.Sum256([]byte(h.AuditionLine))
		sha := hex.EncodeToString(sum[:])
		for _, v := range h.ElevenLabs {
			var exists bool
			err := pool.QueryRow(ctx, `
				SELECT EXISTS (SELECT 1 FROM voice_auditions
				WHERE host = $1 AND vendor = 'elevenlabs' AND model = $2 AND voice_id = $3 AND line_sha256 = $4)`,
				h.Slug, o.Model, v.ID, sha).Scan(&exists)
			if err != nil {
				return err
			}
			if exists {
				skipped++
				continue
			}
			todo = append(todo, job{h, v, sha})
			chars += len([]rune(h.AuditionLine))
		}
	}
	fmt.Fprintf(o.Out, "%d to render (%d characters on %s), %d already done\n", len(todo), chars, o.Model, skipped)
	if o.DryRun {
		for _, j := range todo {
			fmt.Fprintf(o.Out, "  %-7s %-9s %s\n", j.host.Slug, j.voice.Pick, j.voice.Name)
		}
		return nil
	}

	for _, j := range todo {
		mp3, err := tts.Speak(ctx, j.voice.ID, o.Model, j.host.AuditionLine)
		if err != nil {
			return err
		}
		secs, err := audio.ProbeBytesSeconds(ctx, mp3, ".mp3")
		if err != nil {
			return err
		}
		key := fmt.Sprintf("auditions/elevenlabs/%s/%s/%s-%s.mp3", o.Model, j.host.Slug, j.voice.ID, j.sha[:8])
		if err := store.Put(ctx, key, "audio/mpeg", bytes.NewReader(mp3), int64(len(mp3))); err != nil {
			return err
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO voice_auditions (host, vendor, model, voice_id, voice_name, pick, line_sha256,
			                             characters, audio_key, audio_bytes, duration_s)
			VALUES ($1, 'elevenlabs', $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			j.host.Slug, o.Model, j.voice.ID, j.voice.Name, j.voice.Pick, j.sha,
			len([]rune(j.host.AuditionLine)), key, len(mp3), secs)
		if err != nil {
			return fmt.Errorf("record %s/%s (uploaded as %s): %w", j.host.Slug, j.voice.Name, key, err)
		}
		fmt.Fprintf(o.Out, "  %-7s %-9s %-36s %3ds\n", j.host.Slug, j.voice.Pick, j.voice.Name, secs)
	}
	return nil
}
