-- Voice casting auditions: each host's audition line rendered in each shortlisted
-- voice, with your verdict. Audio lives in the bucket under auditions/.

-- +goose Up
CREATE TABLE voice_auditions (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    host        text NOT NULL,             -- slug in show/hosts.yaml
    vendor      text NOT NULL CHECK (vendor IN ('elevenlabs', 'gemini', 'azure')),
    model       text NOT NULL,
    voice_id    text NOT NULL,
    voice_name  text NOT NULL,
    pick        text NOT NULL CHECK (pick IN ('first', 'alternate')),
    line_sha256 text NOT NULL,             -- which version of the audition line
    characters  integer NOT NULL,          -- billed length of the line
    audio_key   text NOT NULL,
    audio_bytes bigint NOT NULL,
    duration_s  integer NOT NULL,
    verdict     text NOT NULL DEFAULT 'undecided'
                CHECK (verdict IN ('undecided', 'shortlist', 'reject')),
    note        text NOT NULL DEFAULT '',
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    UNIQUE (host, vendor, model, voice_id, line_sha256)
);

-- +goose Down
DROP TABLE voice_auditions;
