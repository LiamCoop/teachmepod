-- ADR-001: Topics as a knowledge graph.
-- Concepts, edges and episode coverage are global; what a person knows is per user.

-- +goose Up
CREATE EXTENSION IF NOT EXISTS vector;

-- Listeners. One row today; exists so per-person knowledge is keyed by user from day one.
CREATE TABLE users (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    handle     text NOT NULL UNIQUE,
    name       text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- Global, user-independent nodes.
CREATE TABLE concepts (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    slug        text NOT NULL UNIQUE,
    name        text NOT NULL,
    summary     text NOT NULL DEFAULT '',
    embedding   vector,  -- dimension fixed (and an HNSW index added) once the embedding model is chosen
    status      text NOT NULL DEFAULT 'proposed'
                CHECK (status IN ('proposed', 'approved', 'rejected', 'merged')),
    merged_into bigint REFERENCES concepts (id),
    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now(),
    CHECK ((status = 'merged') = (merged_into IS NOT NULL)),
    CHECK (merged_into <> id)
);

-- Typed links. Direction:
--   prerequisite  src must be learned before dst
--   part_of       src is a child of dst (the tree view)
--   related       lateral; stored once with src_id < dst_id
CREATE TABLE concept_edges (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    src_id     bigint NOT NULL REFERENCES concepts (id) ON DELETE CASCADE,
    dst_id     bigint NOT NULL REFERENCES concepts (id) ON DELETE CASCADE,
    kind       text NOT NULL CHECK (kind IN ('prerequisite', 'part_of', 'related')),
    source     text NOT NULL CHECK (source IN ('pipeline', 'manual')),
    status     text NOT NULL DEFAULT 'proposed'
               CHECK (status IN ('proposed', 'approved', 'rejected')),
    reason     text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (src_id, dst_id, kind),
    CHECK (src_id <> dst_id),
    CHECK (kind <> 'related' OR src_id < dst_id)
);
-- At most one parent per concept, so part_of forms a tree.
CREATE UNIQUE INDEX concept_edges_one_parent ON concept_edges (src_id)
    WHERE kind = 'part_of' AND status <> 'rejected';
CREATE INDEX concept_edges_dst_idx ON concept_edges (dst_id, kind);

-- Reject an edge that would close a cycle in the prerequisite graph or the part_of tree.
-- +goose StatementBegin
CREATE FUNCTION concept_edges_check_acyclic() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.kind NOT IN ('prerequisite', 'part_of') OR NEW.status = 'rejected' THEN
        RETURN NEW;
    END IF;
    -- Serialize graph writes so two concurrent inserts can't form a cycle together.
    PERFORM pg_advisory_xact_lock(hashtext('concept_edges_acyclic'));
    -- A cycle exists if dst already reaches src along edges of the same kind.
    IF EXISTS (
        WITH RECURSIVE reach (id) AS (
            SELECT e.dst_id FROM concept_edges e
            WHERE e.src_id = NEW.dst_id AND e.kind = NEW.kind
              AND e.status <> 'rejected' AND e.id IS DISTINCT FROM NEW.id
            UNION
            SELECT e.dst_id FROM concept_edges e JOIN reach r ON e.src_id = r.id
            WHERE e.kind = NEW.kind AND e.status <> 'rejected' AND e.id IS DISTINCT FROM NEW.id
        )
        SELECT 1 FROM reach WHERE id = NEW.src_id
    ) THEN
        RAISE EXCEPTION 'concept edge % -> % (%) would create a cycle',
            NEW.src_id, NEW.dst_id, NEW.kind
            USING ERRCODE = 'check_violation';
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER concept_edges_acyclic
    BEFORE INSERT OR UPDATE OF src_id, dst_id, kind, status ON concept_edges
    FOR EACH ROW EXECUTE FUNCTION concept_edges_check_acyclic();

-- What each episode covers.
CREATE TABLE episode_concepts (
    episode_id bigint NOT NULL REFERENCES episodes (id) ON DELETE CASCADE,
    concept_id bigint NOT NULL REFERENCES concepts (id),
    role       text NOT NULL CHECK (role IN ('introduces', 'deepens', 'mentions')),
    PRIMARY KEY (episode_id, concept_id)
);
CREATE INDEX episode_concepts_concept_idx ON episode_concepts (concept_id);

-- What a given person knows.
CREATE TABLE user_concepts (
    user_id        bigint NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    concept_id     bigint NOT NULL REFERENCES concepts (id),
    state          text NOT NULL CHECK (state IN ('unknown', 'heard', 'known')),
    via_episode_id bigint REFERENCES episodes (id) ON DELETE SET NULL,
    updated_at     timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (user_id, concept_id)
);

-- A backlog item becomes a request to fill part of the graph.
ALTER TABLE topics ADD COLUMN target_concept_id bigint REFERENCES concepts (id);
CREATE INDEX topics_target_concept_idx ON topics (target_concept_id);

-- +goose Down
ALTER TABLE topics DROP COLUMN target_concept_id;
DROP TABLE user_concepts;
DROP TABLE episode_concepts;
DROP TRIGGER concept_edges_acyclic ON concept_edges;
DROP FUNCTION concept_edges_check_acyclic();
DROP TABLE concept_edges;
DROP TABLE concepts;
DROP TABLE users;
