DROP INDEX IF EXISTS idx_posts_group_id;

ALTER TABLE posts
DROP COLUMN group_id;
