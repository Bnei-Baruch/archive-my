------------
-- Tables --
------------

DROP TABLE IF EXISTS users;
CREATE TABLE users
(
    id          BIGSERIAL PRIMARY KEY,
    accounts_id VARCHAR(36) UNIQUE                     NOT NULL,
    email       VARCHAR(256),
    first_name  VARCHAR(256),
    last_name   VARCHAR(256),
    disabled    BOOLEAN                  DEFAULT FALSE NOT NULL,
    properties  JSONB,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
    updated_at  TIMESTAMP WITH TIME ZONE,
    removed_at  TIMESTAMP WITH TIME ZONE
);

DROP TABLE IF EXISTS playlists;
CREATE TABLE playlists
(
    id         BIGSERIAL PRIMARY KEY,
    uid        CHAR(8) UNIQUE                         NOT NULL,
    user_id    BIGINT REFERENCES users                NOT NULL,
    name       VARCHAR(256),
    public     BOOLEAN                  DEFAULT FALSE NOT NULL,
    properties JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);


DROP TABLE IF EXISTS playlist_items;
CREATE TABLE playlist_items
(
    id               BIGSERIAL PRIMARY KEY,
    playlist_id      BIGINT REFERENCES playlists ON DELETE CASCADE NOT NULL,
    position         INTEGER                                       NOT NULL,
    content_unit_uid CHAR(8)                                       NOT NULL
);


DROP TABLE IF EXISTS reactions;
CREATE TABLE reactions
(
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT REFERENCES users NOT NULL,
    kind         VARCHAR(16)             NOT NULL,
    subject_type VARCHAR(16)             NOT NULL,
    subject_uid  CHAR(8)                 NOT NULL,
    UNIQUE (user_id, kind, subject_type, subject_uid)
);


DROP TABLE IF EXISTS subscriptions;
CREATE TABLE subscriptions
(
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT REFERENCES users                NOT NULL,
    collection_uid   CHAR(8),
    content_type     VARCHAR(32),
    content_unit_uid CHAR(8),
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL,
    updated_at       TIMESTAMP WITH TIME ZONE,
    UNIQUE (user_id, collection_uid),
    UNIQUE (user_id, content_type)
);


DROP TABLE IF EXISTS history;
CREATE TABLE history
(
    id               BIGSERIAL PRIMARY KEY,
    user_id          BIGINT REFERENCES users                NOT NULL,
    chronicle_id     CHAR(27)                               NOT NULL,
    content_unit_uid CHAR(8),
    data             JSONB,
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);


-------------
-- Indexes --
-------------

CREATE
INDEX IF NOT EXISTS playlists_user_id_idx
    ON playlists USING BTREE (user_id, created_at);

CREATE
INDEX IF NOT EXISTS playlist_items_playlist_id_idx
    ON playlist_items USING BTREE (playlist_id, position);

CREATE
INDEX IF NOT EXISTS reactions_user_id_idx
    ON reactions USING BTREE (user_id, kind, subject_type, subject_uid);

CREATE
INDEX IF NOT EXISTS reactions_subject_idx
    ON reactions USING BTREE (subject_type, subject_uid, kind);

CREATE
INDEX IF NOT EXISTS subscriptions_user_id_idx
    ON subscriptions USING BTREE (user_id, created_at);

CREATE
INDEX IF NOT EXISTS history_user_id_idx
    ON history USING BTREE (user_id, created_at);

CREATE
INDEX IF NOT EXISTS history_content_unit_uid_idx
    ON history USING BTREE (user_id, content_unit_uid);
