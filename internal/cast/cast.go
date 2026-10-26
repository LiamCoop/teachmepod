// Package cast loads the hosts and pairings from show/hosts.yaml.
//
// The cast is versioned in git, not stored in Postgres; database rows refer
// to hosts and pairings by slug.
package cast

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"teachmepod/show"
)

type Host struct {
	Slug           string   `yaml:"slug"`
	Name           string   `yaml:"name"`
	Pronouns       string   `yaml:"pronouns"`
	Role           string   `yaml:"role"` // host or guest
	Background     string   `yaml:"background"`
	InConversation string   `yaml:"in_conversation"`
	HowTheyTalk    string   `yaml:"how_they_talk"`
	Hills          []string `yaml:"hills"`
	BlindSpot      string   `yaml:"blind_spot"`
	NeverSays      string   `yaml:"never_says"`
	VoiceCasting   string   `yaml:"voice_casting"`

	// Voice casting shortlist: a 30-second line in character, and the
	// ElevenLabs voices to hear it in.
	AuditionLine string        `yaml:"audition_line"`
	ElevenLabs   []VoiceOption `yaml:"elevenlabs"`
}

type VoiceOption struct {
	Pick string `yaml:"pick"` // first or alternate
	Name string `yaml:"name"`
	ID   string `yaml:"id"`
	Cast bool   `yaml:"cast"` // the voice this host was cast in
}

// CastVoice returns the ElevenLabs voice the host was cast in, if any.
func (h Host) CastVoice() (VoiceOption, bool) {
	for _, v := range h.ElevenLabs {
		if v.Cast {
			return v, true
		}
	}
	return VoiceOption{}, false
}

type Pairing struct {
	Slug       string   `yaml:"slug"`
	Lead       string   `yaml:"lead"`
	Foil       string   `yaml:"foil"`
	Friction   string   `yaml:"friction"`
	TopicAreas []string `yaml:"topic_areas"`
	RunningBit string   `yaml:"running_bit"`
}

type Cast struct {
	Hosts    []Host    `yaml:"hosts"`
	Pairings []Pairing `yaml:"pairings"`
}

// Load parses and validates the embedded cast.
func Load() (*Cast, error) {
	b, err := show.FS.ReadFile("hosts.yaml")
	if err != nil {
		return nil, err
	}
	var c Cast
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("parse hosts.yaml: %w", err)
	}
	if err := c.validate(); err != nil {
		return nil, fmt.Errorf("hosts.yaml: %w", err)
	}
	return &c, nil
}

func (c *Cast) validate() error {
	seen := map[string]bool{}
	for _, h := range c.Hosts {
		if h.Slug == "" || seen[h.Slug] {
			return fmt.Errorf("host slug %q is empty or duplicated", h.Slug)
		}
		if h.Role != "host" && h.Role != "guest" {
			return fmt.Errorf("host %s: role %q must be host or guest", h.Slug, h.Role)
		}
		firsts, cast := 0, 0
		for _, v := range h.ElevenLabs {
			if v.Cast {
				cast++
			}
			if v.ID == "" || v.Name == "" || (v.Pick != "first" && v.Pick != "alternate") {
				return fmt.Errorf("host %s: elevenlabs voice %+v needs an id, a name and pick first or alternate", h.Slug, v)
			}
			if v.Pick == "first" {
				firsts++
			}
		}
		if cast > 1 {
			return fmt.Errorf("host %s: cast in %d voices, at most one allowed", h.Slug, cast)
		}
		if len(h.ElevenLabs) > 0 && firsts != 1 {
			return fmt.Errorf("host %s: elevenlabs needs exactly one first pick, has %d", h.Slug, firsts)
		}
		if len(h.ElevenLabs) > 0 && h.AuditionLine == "" {
			return fmt.Errorf("host %s: elevenlabs voices but no audition_line", h.Slug)
		}
		seen[h.Slug] = true
	}
	pairs := map[string]bool{}
	for _, p := range c.Pairings {
		if p.Slug == "" || pairs[p.Slug] {
			return fmt.Errorf("pairing slug %q is empty or duplicated", p.Slug)
		}
		if !seen[p.Lead] || !seen[p.Foil] || p.Lead == p.Foil {
			return fmt.Errorf("pairing %s: bad lead %q or foil %q", p.Slug, p.Lead, p.Foil)
		}
		pairs[p.Slug] = true
	}
	return nil
}

// Host returns the host with the given slug.
func (c *Cast) Host(slug string) (Host, bool) {
	for _, h := range c.Hosts {
		if h.Slug == slug {
			return h, true
		}
	}
	return Host{}, false
}

// Pairing returns the pairing with the given slug.
func (c *Cast) Pairing(slug string) (Pairing, bool) {
	for _, p := range c.Pairings {
		if p.Slug == slug {
			return p, true
		}
	}
	return Pairing{}, false
}
