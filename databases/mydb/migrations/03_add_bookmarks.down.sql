DROP
INDEX IF EXISTS bookmark_id_folder_id_idx;

DROP
INDEX IF EXISTS bookmarks_user_id_source_idx;

DROP
INDEX IF EXISTS folders_user_id_idx;


DROP TABLE IF EXISTS bookmark;
DROP TABLE IF EXISTS folder;
DROP TABLE IF EXISTS bookmark_folder;
