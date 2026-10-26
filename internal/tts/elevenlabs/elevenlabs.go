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

// Speak renders text in one voice and returns MP3 bytes.
func (c *Client) Speak(ctx context.Context, voiceID, model, text string) ([]byte, error) {
	body, err := json.Marshal(map[string]string{"text": text, "model_id": model})
	if err != nil {
		return nil, err
	}
	u := fmt.Sprintf("%s/v1/text-to-speech/%s?output_format=%s", baseURL, url.PathEscape(voiceID), outputFormat)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("xi-api-key", c.key)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "audio/mpeg")

	resp, err := c.http.Do(req)
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
