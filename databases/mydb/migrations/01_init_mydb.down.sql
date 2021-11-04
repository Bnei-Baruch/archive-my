DROP
    INDEX IF EXISTS playlists_user_id_idx;
DROP
    INDEX IF EXISTS playlist_item_playlist_id_idx;
DROP
    INDEX IF EXISTS reactions_user_id_idx;
DROP
    INDEX IF EXISTS reactions_subject_idx;
DROP
    INDEX IF EXISTS subscriptions_user_id_idx;
DROP
    INDEX IF EXISTS history_user_id_idx;
DROP
    INDEX IF EXISTS history_content_unit_uid_idx;


DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS playlists;
DROP TABLE IF EXISTS playlist_items;
DROP TABLE IF EXISTS reactions;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS history;
DROP TABLE IF EXISTS bookmark;
DROP TABLE IF EXISTS folder;
DROP TABLE IF EXISTS bookmark_folder;
