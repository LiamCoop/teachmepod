package web

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"errors"
	"html/template"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"

	"teachmepod/internal/cast"
)

//go:embed templates/*.html
var templateFS embed.FS

var auditionsTmpl = template.Must(template.ParseFS(templateFS, "templates/auditions.html"))

type clip struct {
	ID        int64
	VoiceName string
	Pick      string
	Model     string
	DurationS int
	Verdict   string
	Note      string
}

type hostAuditions struct {
	cast.Host
	Clips []clip
}

type pairingAuditions struct {
	Title    string
	Friction string
	Hosts    []hostAuditions
}

// auditions serves /auditions/<token>: every host's audition clips, grouped by
// pairing so each pair is heard side by side, with a verdict per clip. Until
// passkey login exists it shares the feed token.
func (s *Server) auditions(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if !s.validToken(token) {
		http.NotFound(w, r)
		return
	}
	byHost := map[string][]clip{}
	rows, err := s.pool.Query(r.Context(), `
		SELECT id, host, voice_name, pick, model, duration_s, verdict, note, line_sha256
		FROM voice_auditions
		ORDER BY host, pick = 'alternate', voice_name, created_at DESC`)
	if err != nil {
		serverError(w, "auditions: query", err)
		return
	}
	current := map[string]string{} // host -> sha of today's audition line
	for _, h := range s.cast.Hosts {
		sum := sha256.Sum256([]byte(h.AuditionLine))
		current[h.Slug] = hex.EncodeToString(sum[:])
	}
	for rows.Next() {
		var c clip
		var host, sha string
		if err := rows.Scan(&c.ID, &host, &c.VoiceName, &c.Pick, &c.Model, &c.DurationS, &c.Verdict, &c.Note, &sha); err != nil {
			serverError(w, "auditions: scan", err)
			return
		}
		if sha == current[host] { // clips of an older line are kept but not shown
			byHost[host] = append(byHost[host], c)
		}
	}
	if err := rows.Err(); err != nil {
		serverError(w, "auditions: rows", err)
		return
	}

	withClips := func(slug string) hostAuditions {
		h, _ := s.cast.Host(slug)
		return hostAuditions{Host: h, Clips: byHost[slug]}
	}
	var groups []pairingAuditions
	for _, p := range s.cast.Pairings {
		lead, _ := s.cast.Host(p.Lead)
		foil, _ := s.cast.Host(p.Foil)
		groups = append(groups, pairingAuditions{
			Title:    lead.Name + " + " + foil.Name,
			Friction: p.Friction,
			Hosts:    []hostAuditions{withClips(p.Lead), withClips(p.Foil)},
		})
	}
	guests := pairingAuditions{Title: "Guests", Friction: "Rotate in alongside any pair"}
	for _, h := range s.cast.Hosts {
		if h.Role == "guest" {
			guests.Hosts = append(guests.Hosts, withClips(h.Slug))
		}
	}
	groups = append(groups, guests)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	if err := auditionsTmpl.Execute(w, map[string]any{"Token": token, "Groups": groups}); err != nil {
		slog.Error("auditions: render", "err", err)
	}
}

// auditionAudio redirects to a presigned URL for one clip.
func (s *Server) auditionAudio(w http.ResponseWriter, r *http.Request) {
	if !s.validToken(r.PathValue("token")) {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	var key string
	err = s.pool.QueryRow(r.Context(), `SELECT audio_key FROM voice_auditions WHERE id = $1`, id).Scan(&key)
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		serverError(w, "audition audio: query", err)
		return
	}
	u, err := s.store.PresignGet(r.Context(), key, audioLinkTTL)
	if err != nil {
		serverError(w, "audition audio: presign", err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	http.Redirect(w, r, u, http.StatusFound)
}

// auditionVerdict records a verdict and note for one clip, then returns to the
// host's section of the page.
func (s *Server) auditionVerdict(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if !s.validToken(token) {
		http.NotFound(w, r)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	verdict := r.FormValue("verdict")
	switch verdict {
	case "undecided", "shortlist", "reject":
	default:
		http.Error(w, "verdict must be undecided, shortlist or reject", http.StatusBadRequest)
		return
	}
	var host string
	err = s.pool.QueryRow(r.Context(), `
		UPDATE voice_auditions SET verdict = $2, note = $3, updated_at = now()
		WHERE id = $1 RETURNING host`, id, verdict, r.FormValue("note")).Scan(&host)
	if errors.Is(err, pgx.ErrNoRows) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		serverError(w, "audition verdict: update", err)
		return
	}
	http.Redirect(w, r, "/auditions/"+token+"#"+host, http.StatusSeeOther)
}
