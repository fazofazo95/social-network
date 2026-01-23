CREATE TABLE IF NOT EXISTS reactions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	target_type TEXT NOT NULL CHECK(target_type IN ('post', 'picture', 'comment')),
	target_id INTEGER NOT NULL,
	reaction_type TEXT NOT NULL CHECK(reaction_type IN ('like', 'dislike')),
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);