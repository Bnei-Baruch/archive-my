DROP
INDEX IF EXISTS bookmark_public_idx;

DROP TABLE IF EXISTS bookmark_tag;

ALTER TABLE bookmarks
DROP
COLUMN public,
    DROP
COLUMN accepted,
    DROP
COLUMN uid;

