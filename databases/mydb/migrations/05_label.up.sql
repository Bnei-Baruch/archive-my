------------
-- Tables --
------------

DROP TABLE IF EXISTS labels;
CREATE TABLE labels
(
    id           BIGSERIAL PRIMARY KEY,
    uid          CHAR(8) UNIQUE                         NOT NULL,
    name         VARCHAR(256),
    user_id      BIGINT REFERENCES users                NOT NULL,
    subject_uid  CHAR(8)                                NOT NULL,
    subject_type VARCHAR(32)                            NOT NULL,
    properties   JSONB,
    accepted     BOOLEAN,
    language     CHAR(2)                                NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);

DROP TABLE IF EXISTS label_tag;
CREATE TABLE label_tag
(
    label_id BIGINT REFERENCES labels ON DELETE CASCADE NOT NULL,
    tag_uid  CHAR(8)                                    NOT NULL,
    PRIMARY KEY (tag_uid, label_id)
);

-------------
-- Indexes --
-------------


CREATE
    INDEX IF NOT EXISTS subject_type_subject_uid_label_uid_idx
    ON labels USING BTREE (subject_type, subject_uid, uid);