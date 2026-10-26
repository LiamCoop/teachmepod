-- RSS enclosures must state their size in bytes.

-- +goose Up
ALTER TABLE renders ADD COLUMN audio_bytes bigint;

-- +goose Down
ALTER TABLE renders DROP COLUMN audio_bytes;
