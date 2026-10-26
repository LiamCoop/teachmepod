// Package auditions renders each host's audition line in every shortlisted
// voice from show/hosts.yaml and stores the clips for review. A clip that
// already exists for the same host, voice, model, text, settings and take is
// never paid for twice.
//
// Beyond the plain line, a run can test variants: extra takes (v3 varies a lot
// from one generation to the next), a stability setting, or a reworded line
// such as one with an audio tag. Variants carry a label on the review page.
package auditions

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"teachmepod/internal/audio"
	"teachmepod/internal/cast"
	"teachmepod/internal/storage"
	"teachmepod/internal/tts/elevenlabs"
)

type Options struct {
	Host      string // only this host; empty for all
	Voice     string // only voices whose ID or name starts with this (case-insensitive)
	Cast      bool   // only each host's cast voice
	Model     string
	Takes     int    // render takes 1..Takes (default 1)
	Stability string // creative, natural or robust; empty for the voice's default
	Line      string // replaces the audition line (needs Host and Label)
	Label     string // what the variant tests, shown on the review page
	DryRun    bool   // list what would be rendered and its size, spend nothing
	Out       io.Writer
}

type job struct {
	host  cast.Host
	voice cast.VoiceOption
	text  string
	sha   string
	take  int
}

// Generate renders every missing audition. tts may be nil when DryRun is set.
func Generate(ctx context.Context, pool *pgxpool.Pool, store *storage.Store, tts *elevenlabs.Client, c *cast.Cast, o Options) error {
	if o.Host != "" {
		if _, ok := c.Host(o.Host); !ok {
			return fmt.Errorf("unknown host %q", o.Host)
		}
	}
	if o.Line != "" && (o.Host == "" || o.Label == "") {
		return errors.New("a replacement line needs -host and -label")
	}
	if o.Takes < 1 {
		o.Takes = 1
	}
	var settings elevenlabs.Settings
	if o.Stability != "" {
		v, ok := elevenlabs.Stabilities[o.Stability]
		if !ok {
			return fmt.Errorf("stability %q must be creative, natural or robust", o.Stability)
		}
		settings.Stability = &v
	}
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	label := o.Label
	if label == "" && o.Stability != "" {
		label = "stability: " + o.Stability
	}

	var todo []job
	var skipped, chars int
	for _, h := range c.Hosts {
		if o.Host != "" && h.Slug != o.Host {
			continue
		}
		text := h.AuditionLine
		if o.Line != "" {
			text = o.Line
		}
		sum := sha256.Sum256([]byte(text))
		sha := hex.EncodeToString(sum[:])
		for _, v := range h.ElevenLabs {
			if o.Cast && !v.Cast {
				continue
			}
			if o.Voice != "" && !matchesVoice(v, o.Voice) {
				continue
			}
			for take := 1; take <= o.Takes; take++ {
				var exists bool
				err := pool.QueryRow(ctx, `
					SELECT EXISTS (SELECT 1 FROM voice_auditions
					WHERE host = $1 AND vendor = 'elevenlabs' AND model = $2 AND voice_id = $3
					  AND line_sha256 = $4 AND settings = $5 AND take = $6)`,
					h.Slug, o.Model, v.ID, sha, settingsJSON, take).Scan(&exists)
				if err != nil {
					return err
				}
				if exists {
					skipped++
					continue
				}
				todo = append(todo, job{h, v, text, sha, take})
				chars += len([]rune(text))
			}
		}
	}
	fmt.Fprintf(o.Out, "%d to render (%d characters on %s), %d already done\n", len(todo), chars, o.Model, skipped)
	if o.DryRun {
		for _, j := range todo {
			fmt.Fprintf(o.Out, "  %-7s %-9s %-36s take %d %s\n", j.host.Slug, j.voice.Pick, j.voice.Name, j.take, label)
		}
		return nil
	}

	variant := ""
	if string(settingsJSON) != "{}" || o.Line != "" {
		sum := sha256.Sum256(append(settingsJSON, o.Line...))
		variant = "-v" + hex.EncodeToString(sum[:4])
	}
	for _, j := range todo {
		mp3, err := tts.Speak(ctx, j.voice.ID, o.Model, j.text, settings)
		if err != nil {
			return err
		}
		secs, err := audio.ProbeBytesSeconds(ctx, mp3, ".mp3")
		if err != nil {
			return err
		}
		key := fmt.Sprintf("auditions/elevenlabs/%s/%s/%s-%s%s-t%d.mp3", o.Model, j.host.Slug, j.voice.ID, j.sha[:8], variant, j.take)
		if err := store.Put(ctx, key, "audio/mpeg", bytes.NewReader(mp3), int64(len(mp3))); err != nil {
			return err
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO voice_auditions (host, vendor, model, voice_id, voice_name, pick, line_sha256,
			                             characters, audio_key, audio_bytes, duration_s, take, settings, label)
			VALUES ($1, 'elevenlabs', $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
			j.host.Slug, o.Model, j.voice.ID, j.voice.Name, j.voice.Pick, j.sha,
			len([]rune(j.text)), key, len(mp3), secs, j.take, settingsJSON, label)
		if err != nil {
			return fmt.Errorf("record %s/%s (uploaded as %s): %w", j.host.Slug, j.voice.Name, key, err)
		}
		fmt.Fprintf(o.Out, "  %-7s %-9s %-36s take %d %3ds %s\n", j.host.Slug, j.voice.Pick, j.voice.Name, j.take, secs, label)
	}
	return nil
}

func matchesVoice(v cast.VoiceOption, prefix string) bool {
	p := strings.ToLower(prefix)
	return strings.HasPrefix(strings.ToLower(v.ID), p) || strings.HasPrefix(strings.ToLower(v.Name), p)
}
