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
	for _, h := range c.Hosts {
		if h.AuditionLine == "" || len(h.ElevenLabs) == 0 {
			t.Errorf("%s: no audition line or voices", h.Slug)
		}
	}
}
