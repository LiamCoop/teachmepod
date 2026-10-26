package cast

import "testing"

func TestLoad(t *testing.T) {
	c, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Hosts) != 10 || len(c.Pairings) != 4 {
		t.Fatalf("got %d hosts, %d pairings; want 10 and 4", len(c.Hosts), len(c.Pairings))
	}
	voiceOf := map[string]string{}
	for _, h := range c.Hosts {
		if h.AuditionLine == "" || len(h.ElevenLabs) == 0 {
			t.Errorf("%s: no audition line or voices", h.Slug)
		}
		v, ok := h.CastVoice()
		if !ok {
			t.Errorf("%s: not cast", h.Slug)
			continue
		}
		if other, taken := voiceOf[v.ID]; taken {
			t.Errorf("%s and %s are both cast as %s", other, h.Slug, v.Name)
		}
		voiceOf[v.ID] = h.Slug
	}
}
