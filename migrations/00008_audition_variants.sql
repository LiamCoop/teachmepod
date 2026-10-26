-- Audition variants: several takes of the same line, different voice settings,
-- or a reworded line (e.g. with an audio tag), labelled so they can be compared.

-- +goose Up
ALTER TABLE voice_auditions
    ADD COLUMN take     integer NOT NULL DEFAULT 1 CHECK (take >= 1),
    ADD COLUMN settings jsonb   NOT NULL DEFAULT '{}',  -- voice_settings sent to the vendor
    ADD COLUMN label    text    NOT NULL DEFAULT '';    -- what this variant tests; '' for the plain line
ALTER TABLE voice_auditions DROP CONSTRAINT voice_auditions_host_vendor_model_voice_id_line_sha256_key;
ALTER TABLE voice_auditions ADD CONSTRAINT voice_auditions_variant_key
    UNIQUE (host, vendor, model, voice_id, line_sha256, settings, take);

-- +goose Down
DELETE FROM voice_auditions WHERE take > 1 OR settings <> '{}' OR label <> '';
ALTER TABLE voice_auditions DROP CONSTRAINT voice_auditions_variant_key;
ALTER TABLE voice_auditions ADD CONSTRAINT voice_auditions_host_vendor_model_voice_id_line_sha256_key
    UNIQUE (host, vendor, model, voice_id, line_sha256);
ALTER TABLE voice_auditions DROP COLUMN label, DROP COLUMN settings, DROP COLUMN take;
