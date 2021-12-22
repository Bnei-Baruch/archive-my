------------
-- Tables --
------------

DROP TABLE IF EXISTS folders;
CREATE TABLE folders
(
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(256),
    user_id    BIGINT REFERENCES users                NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);

DROP TABLE IF EXISTS bookmarks;
CREATE TABLE bookmarks
(
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(256),
    user_id     BIGINT REFERENCES users                NOT NULL,
    source_uid  CHAR(8)                                NOT NULL,
    source_type VARCHAR(32)                            NOT NULL,
    data        JSONB,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);

DROP TABLE IF EXISTS bookmark_folder;
CREATE TABLE bookmark_folder
(
    bookmark_id BIGINT REFERENCES bookmarks ON DELETE CASCADE NOT NULL,
    folder_id   BIGINT REFERENCES folders ON DELETE CASCADE   NOT NULL,
    position    INTEGER,
    PRIMARY KEY (folder_id, bookmark_id)
);

-------------
-- Indexes --
-------------

CREATE
    INDEX IF NOT EXISTS folders_user_id_idx
    ON folders USING BTREE (user_id);

CREATE
    INDEX IF NOT EXISTS bookmarks_user_id_source_idx
    ON bookmarks USING BTREE (user_id, source_uid, source_type);

CREATE
    INDEX IF NOT EXISTS bookmark_id_folder_id_idx
    ON bookmark_folder USING BTREE (bookmark_id, folder_id);
