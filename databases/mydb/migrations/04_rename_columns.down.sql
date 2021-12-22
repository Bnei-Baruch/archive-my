ALTER TABLE bookmarks
    RENAME COLUMN properties TO data;

ALTER TABLE bookmarks
    RENAME COLUMN subject_uid TO source_uid;

ALTER TABLE bookmarks
    RENAME COLUMN subject_type TO source_type;

ALTER TABLE bookmark_folder
    ADD COLUMN position INTEGER;

ALTER TABLE bookmark_folder
    DROP CONSTRAINT bookmark_folder_bookmark_id_uidx;

ALTER TABLE bookmark_folder
    ADD CONSTRAINT bookmark_folder_pkey PRIMARY KEY (folder_id, bookmark_id);


CREATE
    INDEX IF NOT EXISTS bookmark_id_folder_id_idx
    ON bookmark_folder USING BTREE (bookmark_id, folder_id);