// Package seed loads the initial users, backlog and voice profiles from
// show/seed.yaml. Rows are inserted only when missing, so edits made later
// through the site are never overwritten. The cast is not seeded: it lives
// in show/hosts.yaml (see package cast).
package seed

import (
	"context"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"gopkg.in/yaml.v3"

	"teachmepod/internal/cast"
	"teachmepod/show"
)

type User struct {
	Handle string `yaml:"handle"`
	Name   string `yaml:"name"`
}

type Topic struct {
	Title    string `yaml:"title"`
	Angle    string `yaml:"angle"`
	Baseline string `yaml:"baseline"`
	Depth    string `yaml:"depth"`
	Priority int    `yaml:"priority"`
	Series   string `yaml:"series"`
	Pairing  string `yaml:"pairing"`
}

type VoiceProfile struct {
	Name   string `yaml:"name"`
	Vendor string `yaml:"vendor"`
	Model  string `yaml:"model"`
	Notes  string `yaml:"notes"`
	// CastVoices fills voices (host slug -> voice ID) from the cast voices in hosts.yaml.
	CastVoices bool `yaml:"cast_voices"`
	// Voices overrides hosts' voices (host slug -> voice ID). An ElevenLabs
	// override must be on that host's shortlist, which also supplies its stability.
	Voices map[string]string `yaml:"voices"`
}

type Initial struct {
	Users         []User         `yaml:"users"`
	Topics        []Topic        `yaml:"topics"`
	VoiceProfiles []VoiceProfile `yaml:"voice_profiles"`
}

// Run seeds everything in one transaction.
func Run(ctx context.Context, pool *pgxpool.Pool) error {
	c, err := cast.Load()
	if err != nil {
		return err
	}
	b, err := show.FS.ReadFile("seed.yaml")
	if err != nil {
		return err
	}
	var initial Initial
	if err := yaml.Unmarshal(b, &initial); err != nil {
		return fmt.Errorf("parse seed.yaml: %w", err)
	}
	for _, t := range initial.Topics {
		if _, ok := c.Pairing(t.Pairing); t.Pairing != "" && !ok {
			return fmt.Errorf("topic %q: unknown pairing %q", t.Title, t.Pairing)
		}
	}

	return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		for _, u := range initial.Users {
			_, err := tx.Exec(ctx, `
				INSERT INTO users (handle, name) VALUES ($1, $2)
				ON CONFLICT (handle) DO NOTHING`,
				u.Handle, u.Name)
			if err != nil {
				return fmt.Errorf("user %s: %w", u.Handle, err)
			}
		}

		for _, t := range initial.Topics {
			depth := t.Depth
			if depth == "" {
				depth = "intermediate"
			}
			_, err := tx.Exec(ctx, `
				INSERT INTO topics (title, angle, baseline, depth, priority, series, pairing)
				VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''))
				ON CONFLICT (lower(title)) DO NOTHING`,
				t.Title, t.Angle, t.Baseline, depth, t.Priority, t.Series, t.Pairing)
			if err != nil {
				return fmt.Errorf("topic %q: %w", t.Title, err)
			}
		}

		for _, v := range initial.VoiceProfiles {
			voices := map[string]string{}
			if v.CastVoices {
				if v.Vendor != "elevenlabs" {
					return fmt.Errorf("voice profile %s: cast voices exist only for elevenlabs", v.Name)
				}
				for _, h := range c.Hosts {
					if cv, ok := h.CastVoice(); ok {
						voices[h.Slug] = cv.ID
					}
				}
			}
			for slug, id := range v.Voices {
				h, ok := c.Host(slug)
				if !ok {
					return fmt.Errorf("voice profile %s: unknown host %q", v.Name, slug)
				}
				if v.Vendor == "elevenlabs" && !slices.ContainsFunc(h.ElevenLabs, func(o cast.VoiceOption) bool { return o.ID == id }) {
					return fmt.Errorf("voice profile %s: %s is not on %s's shortlist", v.Name, id, slug)
				}
				voices[slug] = id
			}
			_, err := tx.Exec(ctx, `
				INSERT INTO voice_profiles (name, vendor, model, notes, voices)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (name) DO NOTHING`,
				v.Name, v.Vendor, v.Model, v.Notes, voices)
			if err != nil {
				return fmt.Errorf("voice profile %s: %w", v.Name, err)
			}
		}

		fmt.Printf("seeded %d users, %d topics, %d voice profiles\n",
			len(initial.Users), len(initial.Topics), len(initial.VoiceProfiles))
		return nil
	})
}
