package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"teachmepod/internal/storage"
)

// Bad tokens must be rejected before anything touches the database, so these
// run without one (a nil pool would panic if it were used).
func TestFeedRejectsBadTokens(t *testing.T) {
	s := New(nil, &storage.Store{}, "s3cret")
	for _, path := range []string{
		"/feed/wrong.xml",
		"/feed/s3cret",
		"/feed/s3cret.rss",
		"/feed/.xml",
		"/audio/wrong/1/x.mp3",
		"/audio/s3cret/notanumber/x.mp3",
	} {
		rec := httptest.NewRecorder()
		s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", path, rec.Code)
		}
	}
}

func TestFeedDisabledWithoutToken(t *testing.T) {
	s := New(nil, &storage.Store{}, "")
	rec := httptest.NewRecorder()
	s.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/feed/.xml", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("GET /feed/.xml with no token configured = %d, want 404", rec.Code)
	}
}
