DROP VIEW IF EXISTS active_posts;
DROP VIEW IF EXISTS active_comments;

DROP INDEX IF EXISTS idx_posts_user_deleted;
DROP INDEX IF EXISTS idx_posts_privacy;
DROP INDEX IF EXISTS idx_comments_parent;
DROP INDEX IF EXISTS idx_followers_composite;
DROP INDEX IF EXISTS idx_posts_user_id;
DROP INDEX IF EXISTS idx_comments_user_id;