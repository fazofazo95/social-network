PRAGMA foreign_keys = OFF;

CREATE TABLE IF NOT EXISTS notifications_new (
    id INTEGER PRIMARY KEY,
    recipient_id INTEGER NOT NULL,
    actor_id INTEGER,
    type TEXT NOT NULL CHECK (type IN ('follow_request', 'group_invite', 'group_join_request', 'group_event_created')),
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'accepted', 'rejected', 'read')),
    group_id INTEGER,
    event_id INTEGER,
    content TEXT NOT NULL,
    metadata TEXT NOT NULL DEFAULT '{}',
    seen BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    FOREIGN KEY (recipient_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE SET NULL,
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (event_id) REFERENCES group_events(id) ON DELETE CASCADE
);

INSERT INTO notifications_new (id, recipient_id, actor_id, type, status, group_id, event_id, content, metadata, seen, created_at, updated_at)
SELECT
    id,
    user_id,
    parent_id,
    'follow_request',
    CASE WHEN seen = 1 THEN 'read' ELSE 'pending' END,
    NULL,
    NULL,
    content,
    '{}',
    seen,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM notifications;

DROP TABLE notifications;
ALTER TABLE notifications_new RENAME TO notifications;

CREATE INDEX IF NOT EXISTS idx_notifications_recipient_seen_created
ON notifications(recipient_id, seen, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_recipient_status
ON notifications(recipient_id, status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_notifications_type
ON notifications(type);

PRAGMA foreign_keys = ON;