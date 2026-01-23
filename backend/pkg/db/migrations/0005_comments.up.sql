CREATE TABLE IF NOT EXISTS comments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL,
	parent_type TEXT NOT NULL CHECK (parent_type IN ('post', 'picture')),
	parent_id INTEGER NOT NULL,
	content TEXT NOT NULL,
	extra_content TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,

	FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);