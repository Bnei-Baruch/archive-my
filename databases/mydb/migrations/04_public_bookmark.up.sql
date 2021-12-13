------------
-- Tables --
------------

ALTER TABLE bookmarks
    ADD COLUMN public BOOLEAN DEFAULT FALSE NOT NULL,
    ADD COLUMN accepted BOOLEAN DEFAULT NULL,
    ADD COLUMN uid CHAR(8) UNIQUE NOT NULL;


DROP TABLE IF EXISTS bookmark_tag;
CREATE TABLE bookmark_tag
(
    bookmark_id BIGINT REFERENCES bookmarks ON DELETE CASCADE NOT NULL,
    tag_uid     CHAR(8)                                       NOT NULL,
    PRIMARY KEY (tag_uid, bookmark_id)
);


-------------
-- Indexes --
-------------
CREATE
INDEX IF NOT EXISTS bookmark_public_idx
    ON bookmarks USING BTREE (public, uid);