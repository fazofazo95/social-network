-- Migration 0023: allow request_and_invite in join_mode and enforce unique group name

PRAGMA foreign_keys = OFF;

CREATE TABLE IF NOT EXISTS groups_new (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    owner_id INTEGER NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'private')),
    join_mode TEXT NOT NULL DEFAULT 'request' CHECK (join_mode IN ('auto', 'request', 'invite', 'request_and_invite')),
    group_picture TEXT,
    group_members INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);

INSERT INTO groups_new (id, name, description, owner_id, visibility, join_mode, group_picture, group_members, created_at, updated_at)
SELECT id, name, description, owner_id, visibility, join_mode, group_picture, group_members, created_at, updated_at
FROM groups;

DROP TABLE groups;
ALTER TABLE groups_new RENAME TO groups;

CREATE INDEX IF NOT EXISTS idx_groups_owner_id ON groups(owner_id);
CREATE INDEX IF NOT EXISTS idx_groups_visibility ON groups(visibility);
CREATE INDEX IF NOT EXISTS idx_groups_created_at ON groups(created_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_groups_name_unique ON groups(name COLLATE NOCASE);

PRAGMA foreign_keys = ON;
