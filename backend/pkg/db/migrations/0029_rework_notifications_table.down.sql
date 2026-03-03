PRAGMA foreign_keys = OFF;

CREATE TABLE IF NOT EXISTS notifications_old (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	parent_id INTEGER NOT NULL,
	seen BOOLEAN NOT NULL DEFAULT 0,
	content TEXT NOT NULL
);

INSERT INTO notifications_old (id, user_id, parent_id, seen, content)
SELECT
    id,
    recipient_id,
    COALESCE(actor_id, recipient_id),
    seen,
    content
FROM notifications;

DROP TABLE notifications;
ALTER TABLE notifications_old RENAME TO notifications;

PRAGMA foreign_keys = ON;