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
	Stability string // creative, natural or robust; empty for the voice's own from hosts.yaml
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
	v     variant
}

// variant is the settings a clip is rendered with and how the review page labels it.
type variant struct {
	settings elevenlabs.Settings
	json     []byte
	label    string
	suffix   string // added to the bucket key; empty for a plain clip
}

func newVariant(stability, line, label string) (variant, error) {
	var v variant
	if stability != "" {
		s, ok := elevenlabs.Stabilities[stability]
		if !ok {
			return v, fmt.Errorf("stability %q must be creative, natural or robust", stability)
		}
		v.settings.Stability = &s
	}
	var err error
	if v.json, err = json.Marshal(v.settings); err != nil {
		return v, err
	}
	v.label = label
	if v.label == "" && stability != "" {
		v.label = "stability: " + stability
	}
	if string(v.json) != "{}" || line != "" {
		sum := sha256.Sum256(append(v.json, line...))
		v.suffix = "-v" + hex.EncodeToString(sum[:4])
	}
	return v, nil
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
	if _, err := newVariant(o.Stability, "", ""); err != nil {
		return err
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
			stability := o.Stability
			if stability == "" {
				stability = v.Stability
			}
			vr, err := newVariant(stability, o.Line, o.Label)
			if err != nil {
				return fmt.Errorf("%s/%s: %w", h.Slug, v.Name, err)
			}
			for take := 1; take <= o.Takes; take++ {
				var exists bool
				err := pool.QueryRow(ctx, `
					SELECT EXISTS (SELECT 1 FROM voice_auditions
					WHERE host = $1 AND vendor = 'elevenlabs' AND model = $2 AND voice_id = $3
					  AND line_sha256 = $4 AND settings = $5 AND take = $6)`,
					h.Slug, o.Model, v.ID, sha, vr.json, take).Scan(&exists)
				if err != nil {
					return err
				}
				if exists {
					skipped++
					continue
				}
				todo = append(todo, job{h, v, text, sha, take, vr})
				chars += len([]rune(text))
			}
		}
	}
	fmt.Fprintf(o.Out, "%d to render (%d characters on %s), %d already done\n", len(todo), chars, o.Model, skipped)
	if o.DryRun {
		for _, j := range todo {
			fmt.Fprintf(o.Out, "  %-7s %-9s %-36s take %d %s\n", j.host.Slug, j.voice.Pick, j.voice.Name, j.take, j.v.label)
		}
		return nil
	}

	for _, j := range todo {
		mp3, err := tts.Speak(ctx, j.voice.ID, o.Model, j.text, j.v.settings)
		if err != nil {
			return err
		}
		secs, err := audio.ProbeBytesSeconds(ctx, mp3, ".mp3")
		if err != nil {
			return err
		}
		key := fmt.Sprintf("auditions/elevenlabs/%s/%s/%s-%s%s-t%d.mp3", o.Model, j.host.Slug, j.voice.ID, j.sha[:8], j.v.suffix, j.take)
		if err := store.Put(ctx, key, "audio/mpeg", bytes.NewReader(mp3), int64(len(mp3))); err != nil {
			return err
		}
		_, err = pool.Exec(ctx, `
			INSERT INTO voice_auditions (host, vendor, model, voice_id, voice_name, pick, line_sha256,
			                             characters, audio_key, audio_bytes, duration_s, take, settings, label)
			VALUES ($1, 'elevenlabs', $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
			j.host.Slug, o.Model, j.voice.ID, j.voice.Name, j.voice.Pick, j.sha,
			len([]rune(j.text)), key, len(mp3), secs, j.take, j.v.json, j.v.label)
		if err != nil {
			return fmt.Errorf("record %s/%s (uploaded as %s): %w", j.host.Slug, j.voice.Name, key, err)
		}
		fmt.Fprintf(o.Out, "  %-7s %-9s %-36s take %d %3ds %s\n", j.host.Slug, j.voice.Pick, j.voice.Name, j.take, secs, j.v.label)
	}
	return nil
}

func matchesVoice(v cast.VoiceOption, prefix string) bool {
	p := strings.ToLower(prefix)
	return strings.HasPrefix(strings.ToLower(v.ID), p) || strings.HasPrefix(strings.ToLower(v.Name), p)
}
