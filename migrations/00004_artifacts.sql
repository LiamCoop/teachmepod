-- Text artifacts (dossiers, outlines, beats, scripts, review reports, show notes,
-- transcripts) live in Postgres, so they're covered by database backups. The bucket
-- holds audio only, which can always be re-rendered from a script.

-- +goose Up
CREATE TABLE artifacts (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    episode_id bigint NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
    kind       text NOT NULL CHECK (kind IN ('dossier', 'outline', 'beats', 'script',
                                             'review', 'show_notes', 'transcript')),
    media_type text NOT NULL CHECK (media_type IN ('application/json', 'text/markdown', 'text/plain')),
    body       text NOT NULL,
    sha256     text NOT NULL, -- hex digest of body; feeds downstream stages' input_hash
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX artifacts_episode_idx ON artifacts (episode_id, kind, created_at DESC);

-- Text stages point at the artifact they wrote; render and mix write renders rows instead.
ALTER TABLE stage_runs
    DROP COLUMN output_key,
    ADD COLUMN output_artifact_id bigint REFERENCES artifacts (id);

ALTER TABLE episodes
    DROP COLUMN script_key,
    ADD COLUMN script_artifact_id bigint REFERENCES artifacts (id);

-- +goose Down
ALTER TABLE episodes
    DROP COLUMN script_artifact_id,
    ADD COLUMN script_key text;
ALTER TABLE stage_runs
    DROP COLUMN output_artifact_id,
    ADD COLUMN output_key text;
DROP TABLE artifacts;
