
DROP
INDEX IF EXISTS history_chronicles_timestamp_idx;

ALTER TABLE reactions ALTER COLUMN subject_type TYPE varchar(16);

ALTER TABLE playlists DROP COLUMN poster_unit_uid;