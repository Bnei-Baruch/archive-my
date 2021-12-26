------------
-- Tables --
------------


ALTER TABLE bookmarks
    RENAME COLUMN data TO properties;

ALTER TABLE bookmarks
    RENAME COLUMN source_uid TO subject_uid;

ALTER TABLE bookmarks
    RENAME COLUMN source_type TO subject_type;

DROP
    INDEX IF EXISTS bookmark_id_folder_id_idx;
