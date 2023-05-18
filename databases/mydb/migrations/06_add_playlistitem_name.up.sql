------------
-- Tables --
------------


ALTER TABLE playlist_items
    ADD COLUMN name       VARCHAR(256),
    ADD COLUMN properties JSONB