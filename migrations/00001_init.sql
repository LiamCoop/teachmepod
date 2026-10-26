-- +goose Up

-- Hosts and pairings live in show/hosts.yaml, not here. Rows refer to them by slug,
-- and the app validates slugs against the YAML.

-- The backlog.
CREATE TABLE topics (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title      text NOT NULL,
    angle      text NOT NULL DEFAULT '',
    baseline   text NOT NULL DEFAULT '',   -- what the listener already knows
    depth      text NOT NULL DEFAULT 'intermediate'
               CHECK (depth IN ('intro', 'intermediate', 'advanced')),
    priority   integer NOT NULL DEFAULT 0, -- higher runs first
    status     text NOT NULL DEFAULT 'queued'
               CHECK (status IN ('queued', 'in_progress', 'done', 'parked')),
    series     text,
    pairing    text,                       -- slug in show/hosts.yaml; null: the outline stage picks one
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX topics_title_key ON topics (lower(title));
CREATE INDEX topics_queue_idx ON topics (priority DESC, created_at) WHERE status = 'queued';

-- A vendor + voice pairing. voices maps host slug -> vendor voice ID.
CREATE TABLE voice_profiles (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name       text NOT NULL UNIQUE,
    vendor     text NOT NULL CHECK (vendor IN ('elevenlabs', 'gemini', 'azure')),
    model      text NOT NULL DEFAULT '',
    voices     jsonb NOT NULL DEFAULT '{}',
    settings   jsonb NOT NULL DEFAULT '{}',
    notes      text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- One per topic attempt.
CREATE TABLE episodes (
    id                  bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    topic_id            bigint NOT NULL REFERENCES topics (id),
    pairing             text,              -- slug in show/hosts.yaml
    guest               text,              -- host slug in show/hosts.yaml
    status              text NOT NULL DEFAULT 'queued'
                        CHECK (status IN ('queued', 'researching', 'scripting', 'reviewing',
                                          'rendering', 'published', 'failed', 'parked')),
    status_reason       text NOT NULL DEFAULT '',
    title               text NOT NULL DEFAULT '',
    slug                text UNIQUE,
    script_key          text,
    prompt_version      text NOT NULL DEFAULT '',
    published_render_id bigint, -- FK added below, after renders exists
    published_at        timestamptz,
    created_at          timestamptz NOT NULL DEFAULT now(),
    updated_at          timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX episodes_topic_idx ON episodes (topic_id);

-- Every stage execution.
CREATE TABLE stage_runs (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    episode_id  bigint NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
    stage       text NOT NULL CHECK (stage IN ('research', 'outline', 'beats', 'dialogue',
                                               'review', 'render', 'mix', 'publish')),
    attempt     integer NOT NULL DEFAULT 1,
    status      text NOT NULL DEFAULT 'running'
                CHECK (status IN ('running', 'succeeded', 'failed', 'skipped')),
    input_hash  text NOT NULL DEFAULT '',
    output_key  text,
    tokens_in   integer NOT NULL DEFAULT 0,
    tokens_out  integer NOT NULL DEFAULT 0,
    cost_cents  integer NOT NULL DEFAULT 0,
    error       text,
    started_at  timestamptz NOT NULL DEFAULT now(),
    finished_at timestamptz,
    UNIQUE (episode_id, stage, attempt)
);
CREATE INDEX stage_runs_cache_idx ON stage_runs (stage, input_hash) WHERE status = 'succeeded';

-- One audio rendering of a script.
CREATE TABLE renders (
    id               bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    episode_id       bigint NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
    voice_profile_id bigint NOT NULL REFERENCES voice_profiles (id),
    status           text NOT NULL DEFAULT 'queued'
                     CHECK (status IN ('queued', 'rendering', 'done', 'failed')),
    audio_key        text,
    duration_s       integer,
    chapters         jsonb NOT NULL DEFAULT '[]',
    wer              real,
    cost_cents       integer NOT NULL DEFAULT 0,
    is_test          boolean NOT NULL DEFAULT false,
    created_at       timestamptz NOT NULL DEFAULT now(),
    finished_at      timestamptz
);
CREATE INDEX renders_episode_idx ON renders (episode_id);

ALTER TABLE episodes
    ADD CONSTRAINT episodes_published_render_fk
    FOREIGN KEY (published_render_id) REFERENCES renders (id);

-- Reactions after listening.
CREATE TABLE feedback (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    render_id  bigint NOT NULL REFERENCES renders (id) ON DELETE CASCADE,
    kind       text NOT NULL CHECK (kind IN ('cringe', 'wrong', 'deeper', 'too_basic', 'rating')),
    at_seconds integer,
    rating     smallint CHECK (rating BETWEEN 1 AND 5),
    note       text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK ((kind = 'rating') = (rating IS NOT NULL))
);
CREATE INDEX feedback_render_idx ON feedback (render_id);

-- Facts the hosts can call back to.
CREATE TABLE show_memory (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    episode_id bigint NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
    line       text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX show_memory_episode_idx ON show_memory (episode_id);

-- +goose Down
DROP TABLE show_memory;
DROP TABLE feedback;
ALTER TABLE episodes DROP CONSTRAINT episodes_published_render_fk;
DROP TABLE renders;
DROP TABLE stage_runs;
DROP TABLE episodes;
DROP TABLE voice_profiles;
DROP TABLE topics;
