CREATE TABLE IF NOT EXISTS user_relationships (
    user_id1 INTEGER NOT NULL,
    user_id2 INTEGER NOT NULL,
    status TEXT NOT NULL CHECK(status IN ('pending', 'friends', 'blocked')),
    action_user INTEGER NOT NULL,
    seen BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(user_id1, user_id2),

    FOREIGN KEY(user_id1) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY(user_id2) REFERENCES users(id) ON DELETE CASCADE
);