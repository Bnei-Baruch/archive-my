------------
-- Tables --
------------

ALTER TABLE reactions ALTER COLUMN subject_type TYPE varchar(32);

ALTER TABLE playlists
    ADD COLUMN poster_unit_uid CHAR(8);


-------------
-- Indexes --
-------------
CREATE
INDEX IF NOT EXISTS history_chronicles_timestamp_idx
    ON history USING BTREE (user_id, chronicles_timestamp);