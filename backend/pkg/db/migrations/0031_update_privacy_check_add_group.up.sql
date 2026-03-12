-- Migration 0031: update privacy CHECK constraint to allow 'group' value

-- SQLite does not support ALTER COLUMN, so we must recreate the table.

PRAGMA foreign_keys = OFF;

-- 1. Drop dependent views
DROP VIEW IF EXISTS active_posts;

-- 2. Rename old table
ALTER TABLE posts RENAME TO posts_old;

-- 3. Create new table with updated CHECK constraint
CREATE TABLE posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    content TEXT NOT NULL,
    extra_content TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    privacy TEXT CHECK(privacy IN ('public', 'followers', 'custom', 'group')),
    deleted_at DATETIME DEFAULT NULL,
    like_count INTEGER DEFAULT 0,
    group_id INTEGER REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 4. Copy data
INSERT INTO posts (id, user_id, content, extra_content, created_at, privacy, deleted_at, like_count, group_id)
SELECT id, user_id, content, extra_content, created_at, privacy, deleted_at, like_count, group_id
FROM posts_old;

-- 5. Drop old table
DROP TABLE posts_old;

-- 6. Recreate indexes
CREATE INDEX IF NOT EXISTS idx_posts_user_deleted ON posts(user_id, deleted_at);
CREATE INDEX IF NOT EXISTS idx_posts_privacy ON posts(privacy);
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_group_id ON posts(group_id);

-- 7. Recreate view
CREATE VIEW IF NOT EXISTS active_posts AS
SELECT * FROM posts WHERE deleted_at IS NULL;

PRAGMA foreign_keys = ON;
