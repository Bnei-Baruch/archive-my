------------
-- Tables --
------------


DROP TABLE IF EXISTS notes;
CREATE TABLE notes
(
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT REFERENCES users                NOT NULL,
    content     TEXT,
    subject_uid CHAR(8)                                NOT NULL,
    properties  JSONB,
    language    CHAR(2)                                NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT now() NOT NULL
);

-------------
-- Indexes --
-------------

CREATE
    INDEX IF NOT EXISTS notes_user_id_subject_uid_idx
    ON notes USING BTREE (user_id, subject_uid);