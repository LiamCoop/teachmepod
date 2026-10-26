-- After auditions (Oct 26, 2026): Inès is renamed Claire and Mei is renamed Maya,
-- names TTS can't mispronounce; EL-1 carries the cast's voices; EL-2 is dropped,
-- since both of its Dana + Marcus candidates were rejected.

-- +goose Up
UPDATE topics   SET pairing = 'maya-jules'  WHERE pairing = 'mei-jules';
UPDATE topics   SET pairing = 'claire-theo' WHERE pairing = 'ines-theo';
UPDATE episodes SET pairing = 'maya-jules'  WHERE pairing = 'mei-jules';
UPDATE episodes SET pairing = 'claire-theo' WHERE pairing = 'ines-theo';
UPDATE episodes SET guest = 'maya'   WHERE guest = 'mei';
UPDATE episodes SET guest = 'claire' WHERE guest = 'ines';
UPDATE voice_auditions SET host = 'maya'   WHERE host = 'mei';
UPDATE voice_auditions SET host = 'claire' WHERE host = 'ines';

UPDATE voice_profiles SET
    voices = '{"dana": "g7LVvkPWALzPxOQbF6OE", "marcus": "l7kNoIfnJKPg7779LI2t",
               "maya": "dvbL7qkNGZY1IqPGZAjM", "jules": "QtY3JBOUKEB5xzrRfOKc",
               "claire": "BFd5oBc2DDna33pSi4Gf", "theo": "7QN34D2r3hCNwbOYIeK0",
               "priya": "22N9cF8z0o7y23njdyaY", "kofi": "gOupLcAkjEnguROwi4oS",
               "rachel": "OZ0L6eISlOejga3XjDFt", "sam": "fs2OqxduwXgp9foh2xjK"}',
    notes = 'ElevenLabs v3 Text to Dialogue, the cast from auditions',
    updated_at = now()
WHERE name = 'EL-1';

DELETE FROM voice_profiles p
WHERE p.name = 'EL-2' AND NOT EXISTS (SELECT 1 FROM renders r WHERE r.voice_profile_id = p.id);

-- +goose Down
-- Renames and the EL-1 voices are not reversed: the old names are retired on purpose.
SELECT 1;
