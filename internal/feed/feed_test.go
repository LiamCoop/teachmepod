package feed

import (
	"bytes"
	"encoding/xml"
	"strings"
	"testing"
	"time"
)

func TestWrite(t *testing.T) {
	var b bytes.Buffer
	err := Write(&b, Channel{
		Title: "teachmepod", Link: "https://example.test",
		Items: []Item{{
			GUID: "teachmepod-render-7", Title: "Test: EL-1 · Dana & Marcus",
			AudioURL: "https://example.test/audio/tok/7/x.mp3", AudioBytes: 1234,
			Duration: 3*time.Minute + 5*time.Second, Published: time.Date(2026, 10, 26, 18, 0, 0, 0, time.UTC),
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{
		`<rss version="2.0" xmlns:itunes="http://www.itunes.com/dtds/podcast-1.0.dtd">`,
		`<itunes:block>Yes</itunes:block>`,
		`<title>Test: EL-1 · Dana &amp; Marcus</title>`,
		`<guid isPermaLink="false">teachmepod-render-7</guid>`,
		`<pubDate>Mon, 26 Oct 2026 18:00:00 +0000</pubDate>`,
		`<enclosure url="https://example.test/audio/tok/7/x.mp3" length="1234" type="audio/mpeg"></enclosure>`,
		`<itunes:duration>0:03:05</itunes:duration>`,
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s in:\n%s", want, out)
		}
	}
	if err := xml.Unmarshal(b.Bytes(), new(struct{})); err != nil {
		t.Errorf("not well-formed XML: %v", err)
	}
}
