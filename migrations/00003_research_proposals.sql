-- ADR-001 amendment: research proposes new concepts, which you approve into the graph
-- and, optionally, into the backlog.
--   - concepts and edges record the episode whose research proposed them, and why
--   - acyclicity and the one-parent rule hold for approved edges only, so competing
--     proposals can sit side by side in the queue; both are checked on approval

-- +goose Up
ALTER TABLE concepts
    ADD COLUMN proposed_by_episode_id bigint REFERENCES episodes (id) ON DELETE SET NULL,
    ADD COLUMN proposal_reason        text NOT NULL DEFAULT '';
CREATE INDEX concepts_proposed_by_idx ON concepts (proposed_by_episode_id);

ALTER TABLE concept_edges
    ADD COLUMN proposed_by_episode_id bigint REFERENCES episodes (id) ON DELETE SET NULL;
CREATE INDEX concept_edges_proposed_by_idx ON concept_edges (proposed_by_episode_id);

DROP INDEX concept_edges_one_parent;
CREATE UNIQUE INDEX concept_edges_one_parent ON concept_edges (src_id)
    WHERE kind = 'part_of' AND status = 'approved';

-- Reject approving an edge that would close a cycle among approved edges of the same kind.
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION concept_edges_check_acyclic() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    IF NEW.kind NOT IN ('prerequisite', 'part_of') OR NEW.status <> 'approved' THEN
        RETURN NEW;
    END IF;
    -- Serialize graph writes so two concurrent approvals can't form a cycle together.
    PERFORM pg_advisory_xact_lock(hashtext('concept_edges_acyclic'));
    -- A cycle exists if dst already reaches src along approved edges of the same kind.
    IF EXISTS (
        WITH RECURSIVE reach (id) AS (
            SELECT e.dst_id FROM concept_edges e
            WHERE e.src_id = NEW.dst_id AND e.kind = NEW.kind
              AND e.status = 'approved' AND e.id IS DISTINCT FROM NEW.id
            UNION
            SELECT e.dst_id FROM concept_edges e JOIN reach r ON e.src_id = r.id
            WHERE e.kind = NEW.kind AND e.status = 'approved' AND e.id IS DISTINCT FROM NEW.id
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

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION concept_edges_check_acyclic() RETURNS trigger
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

DROP INDEX concept_edges_one_parent;
CREATE UNIQUE INDEX concept_edges_one_parent ON concept_edges (src_id)
    WHERE kind = 'part_of' AND status <> 'rejected';

DROP INDEX concept_edges_proposed_by_idx;
ALTER TABLE concept_edges DROP COLUMN proposed_by_episode_id;
DROP INDEX concepts_proposed_by_idx;
ALTER TABLE concepts DROP COLUMN proposal_reason, DROP COLUMN proposed_by_episode_id;
