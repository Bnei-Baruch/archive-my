BEGIN;

---------------
-- Functions --
---------------

DROP FUNCTION IF EXISTS now_utc();

CREATE FUNCTION now_utc()
    RETURNS TIMESTAMP AS
$$
SELECT now() AT TIME ZONE 'utc';
$$
    LANGUAGE SQL;


------------
-- Tables --
------------

DROP TABLE IF EXISTS playlist;
CREATE TABLE playlist
(
    id          BIGSERIAL PRIMARY KEY,
    account_id  VARCHAR(64)                                NOT NULL,
    name        CHAR(32)                                   NULL,
    parameters  JSONB,
    public      BOOLEAN,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT now_utc() NOT NULL,
    last_played VARCHAR(8)
);


DROP TABLE IF EXISTS playlist_item;
CREATE TABLE playlist_item
(
    id               BIGSERIAL PRIMARY KEY,
    playlist_id      BIGINT REFERENCES playlist                 NOT NULL,
    position         int,
    content_unit_uid VARCHAR(8)                                 NOT NULL,
    added_at         TIMESTAMP WITH TIME ZONE DEFAULT now_utc() NOT NULL
);


DROP TABLE IF EXISTS likes;
CREATE TABLE likes
(
    id               BIGSERIAL PRIMARY KEY,
    account_id       VARCHAR(64)                                NOT NULL,
    content_unit_uid VARCHAR(8)                                 NOT NULL,
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT now_utc() NOT NULL
);


DROP TABLE IF EXISTS subscriptions;
CREATE TABLE subscriptions
(
    id             BIGSERIAL PRIMARY KEY,
    account_id     VARCHAR(64)                                NOT NULL,
    collection_uid VARCHAR(8)                                 NULL,
    content_type   BIGINT                                     NULL,
    updated_at     TIMESTAMP WITH TIME ZONE DEFAULT now_utc() NOT NULL
);


DROP TABLE IF EXISTS history;
CREATE TABLE history
(
    id               BIGSERIAL PRIMARY KEY,
    account_id       VARCHAR(36)                                NOT NULL,
    chronicle_id     VARCHAR(64)                                NOT NULL,
    content_unit_uid VARCHAR(8)                                 NULL,
    data             JSONB,
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT now_utc() NOT NULL
);


-------------
-- Indexes --
-------------

CREATE
    INDEX IF NOT EXISTS playlist_account_id_idx
    ON playlist USING BTREE (account_id);

CREATE
    INDEX IF NOT EXISTS playlist_item_playlist_id_idx
    ON playlist_item USING BTREE (playlist_id);

CREATE
    INDEX IF NOT EXISTS likes_account_id_idx
    ON likes USING BTREE (account_id);

CREATE
    INDEX IF NOT EXISTS subscriptions_account_id_idx
    ON subscriptions USING BTREE (account_id);

CREATE
    INDEX IF NOT EXISTS history_account_id_idx
    ON history USING BTREE (account_id);

CREATE
    INDEX IF NOT EXISTS history_created_at_idx
    ON history USING BTREE (created_at);

CREATE
    INDEX IF NOT EXISTS history_account_id_content_unit_uid_created_at_idx
    ON history USING BTREE (account_id, content_unit_uid, created_at);


COMMIT;