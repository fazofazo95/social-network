-- Rollback 0023: remove request_and_invite option and drop unique group name index

PRAGMA foreign_keys = OFF;

CREATE TABLE IF NOT EXISTS groups_old (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    owner_id INTEGER NOT NULL,
    visibility TEXT NOT NULL DEFAULT 'public' CHECK (visibility IN ('public', 'private')),
    join_mode TEXT NOT NULL DEFAULT 'request' CHECK (join_mode IN ('auto', 'request', 'invite')),
    group_picture TEXT,
    group_members INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);

INSERT INTO groups_old (id, name, description, owner_id, visibility, join_mode, group_picture, group_members, created_at, updated_at)
SELECT id, name, description, owner_id, visibility,
       CASE WHEN join_mode = 'request_and_invite' THEN 'request' ELSE join_mode END,
       group_picture, group_members, created_at, updated_at
FROM groups;

DROP INDEX IF EXISTS idx_groups_name_unique;
DROP TABLE groups;
ALTER TABLE groups_old RENAME TO groups;

CREATE INDEX IF NOT EXISTS idx_groups_owner_id ON groups(owner_id);
CREATE INDEX IF NOT EXISTS idx_groups_visibility ON groups(visibility);
CREATE INDEX IF NOT EXISTS idx_groups_created_at ON groups(created_at);

PRAGMA foreign_keys = ON;
