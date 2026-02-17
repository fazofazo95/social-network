CREATE INDEX IF NOT EXISTS idx_posts_user_deleted ON posts(user_id, deleted_at);
CREATE INDEX IF NOT EXISTS idx_posts_privacy ON posts(privacy);
CREATE INDEX IF NOT EXISTS idx_comments_parent ON comments(parent_id, parent_type, deleted_at);
CREATE INDEX IF NOT EXISTS idx_followers_composite ON followers(follower_id, followed_id);
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);

CREATE VIEW IF NOT EXISTS active_posts AS
SELECT * FROM posts WHERE deleted_at IS NULL;

CREATE VIEW IF NOT EXISTS active_comments AS
SELECT * FROM comments WHERE deleted_at IS NULL;