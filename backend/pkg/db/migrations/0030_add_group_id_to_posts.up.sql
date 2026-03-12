-- Migration 0030: add group_id to posts table for group posts support

ALTER TABLE posts ADD COLUMN group_id INTEGER REFERENCES groups(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_posts_group_id ON posts(group_id);
