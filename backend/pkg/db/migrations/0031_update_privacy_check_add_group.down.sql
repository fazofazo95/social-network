-- Rollback: revert privacy CHECK constraint to exclude 'group'

PRAGMA foreign_keys = OFF;

DROP VIEW IF EXISTS active_posts;

ALTER TABLE posts RENAME TO posts_old;

CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    extra_content TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    privacy TEXT CHECK(privacy IN ('public', 'followers', 'custom')),
    deleted_at DATETIME DEFAULT NULL,
    like_count INTEGER DEFAULT 0,
    group_id INTEGER REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

INSERT INTO posts (id, user_id, content, extra_content, created_at, privacy, deleted_at, like_count, group_id)
SELECT id, user_id, content, extra_content, created_at, privacy, deleted_at, like_count, group_id
FROM posts_old
WHERE privacy != 'group';

DROP TABLE posts_old;

CREATE INDEX IF NOT EXISTS idx_posts_user_deleted ON posts(user_id, deleted_at);
CREATE INDEX IF NOT EXISTS idx_posts_privacy ON posts(privacy);
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_group_id ON posts(group_id);

CREATE VIEW IF NOT EXISTS active_posts AS
SELECT * FROM posts WHERE deleted_at IS NULL;

PRAGMA foreign_keys = ON;
