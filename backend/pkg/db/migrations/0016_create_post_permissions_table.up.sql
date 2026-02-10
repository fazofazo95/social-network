CREATE TABLE IF NOT EXISTS post_permissions (
	post_id INTEGER,
	user_id INTEGER,

	FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE
    PRIMARY KEY(post_id,user_id)
);