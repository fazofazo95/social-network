DROP VIEW IF EXISTS active_posts;
DROP VIEW IF EXISTS active_comments;

DROP INDEX IF EXISTS idx_posts_user_deleted;
DROP INDEX IF EXISTS idx_posts_privacy;
DROP INDEX IF EXISTS idx_comments_parent;
DROP INDEX IF EXISTS idx_followers_composite;