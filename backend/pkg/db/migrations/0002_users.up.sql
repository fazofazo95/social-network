CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY,
	first_name TEXT NOT NULL,
	last_name TEXT NOT NULL,
	birthday_date DATETIME,
	relationship_status TEXT,
	employed_at TEXT,
	phone_number TEXT,
	profile_picture TEXT,
	pictures TEXT,
	level TEXT NOT NULL,

	FOREIGN KEY (id) REFERENCES login_users(id) ON DELETE CASCADE
);