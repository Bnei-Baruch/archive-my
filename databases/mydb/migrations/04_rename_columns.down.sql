ALTER TABLE bookmarks
    RENAME COLUMN properties TO data;

ALTER TABLE bookmarks
    RENAME COLUMN subject_uid TO source_uid;

ALTER TABLE bookmarks
    RENAME COLUMN subject_type TO source_type;

CREATE
    INDEX IF NOT EXISTS bookmark_id_folder_id_idx
    ON bookmark_folder USING BTREE (bookmark_id, folder_id);