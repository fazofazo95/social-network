CREATE TABLE IF NOT EXISTS followers (
    follower_id INTEGER NOT NULL,
    followed_id INTEGER NOT NULL,
    status TEXT NOT NULL CHECK(STATUS in ('pending', 'accepted', 'blocked')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,

    PRIMARY KEY (follower_id, followed_id),

    CHECK(followed_id != follower_id),

    FOREIGN KEY (follower_id) REFERENCES login_users(id) ON DELETE CASCADE,
    FOREIGN KEY (followed_id) REFERENCES login_users(id) ON DELETE CASCADE
)