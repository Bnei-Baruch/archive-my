BEGIN;

DROP
INDEX IF EXISTS playlist_account_id_idx;
DROP
INDEX IF EXISTS playlist_item_playlist_id_idx;
DROP
INDEX IF EXISTS likes_account_id_idx;
DROP
INDEX IF EXISTS subscriptions_account_id_idx;
DROP
INDEX IF EXISTS history_account_id_idx;
DROP
INDEX IF EXISTS history_created_at_idx;
DROP
INDEX IF EXISTS history_account_id_content_unit_uid_created_at_idx;


DROP TABLE IF EXISTS playlist;
DROP TABLE IF EXISTS playlist_item;
DROP TABLE IF EXISTS likes;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS history;
DROP FUNCTION IF EXISTS now_utc();
COMMIT;