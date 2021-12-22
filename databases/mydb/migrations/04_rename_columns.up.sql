------------
-- Tables --
------------


ALTER TABLE bookmarks
    RENAME COLUMN data TO properties;

ALTER TABLE bookmarks
    RENAME COLUMN source_uid TO subject_uid;

ALTER TABLE bookmarks
    RENAME COLUMN source_type TO subject_type;

ALTER TABLE bookmarks
    ADD COLUMN position INTEGER;


ALTER TABLE bookmark_folder
    DROP CONSTRAINT bookmark_folder_bookmark_id_fkey;

ALTER TABLE bookmark_folder
    ADD CONSTRAINT bookmark_folder_bookmark_id_uidx UNIQUE (folder_id, bookmark_id);

DROP
    INDEX IF EXISTS folders_user_id_idx;
