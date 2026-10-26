// Package elevenlabs calls the ElevenLabs text-to-speech API. So far it does
// single-voice speech for auditions; dialogue mode arrives with the bake-off.
package elevenlabs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

const (
	baseURL = "https://api.elevenlabs.io"
	// DefaultModel is the model episodes are expected to use, so auditions use it too.
	DefaultModel = "eleven_v3"
	outputFormat = "mp3_44100_128"
)

type Client struct {
	key  string
	http *http.Client
}

// FromEnv reads ELEVENLABS_API_KEY.
func FromEnv() (*Client, error) {
	key := os.Getenv("ELEVENLABS_API_KEY")
	if key == "" {
		return nil, errors.New("ELEVENLABS_API_KEY is not set")
	}
	return &Client{key: key, http: &http.Client{Timeout: 3 * time.Minute}}, nil
}

// Settings are the voice_settings sent with a request. Nil fields use the
// voice's defaults.
type Settings struct {
	// Stability on eleven_v3: 0 is Creative (most expressive), 0.5 Natural
	// (the default), 1 Robust (steadiest, least responsive to audio tags).
	Stability *float64 `json:"stability,omitempty"`
}

// Stabilities maps the names ElevenLabs uses for v3 stability to values.
var Stabilities = map[string]float64{"creative": 0, "natural": 0.5, "robust": 1}

// Speak renders text in one voice and returns MP3 bytes.
func (c *Client) Speak(ctx context.Context, voiceID, model, text string, settings Settings) ([]byte, error) {
	req := map[string]any{"text": text, "model_id": model}
	if settings != (Settings{}) {
		req["voice_settings"] = settings
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("%s/v1/text-to-speech/%s?output_format=%s", baseURL, url.PathEscape(voiceID), outputFormat)
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	hreq.Header.Set("xi-api-key", c.key)
	hreq.Header.Set("Content-Type", "application/json")
	hreq.Header.Set("Accept", "audio/mpeg")

	resp, err := c.http.Do(hreq)
	if err != nil {
		return nil, fmt.Errorf("elevenlabs: %w", err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("elevenlabs: read: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		if len(b) > 500 {
			b = b[:500]
		}
		return nil, fmt.Errorf("elevenlabs: voice %s: %s: %s", voiceID, resp.Status, b)
	}
	return b, nil
}
