-- Migration 0019: remove obsolete columns from user_settings

-- Recreate the user_settings table without posts_vis, created_date_vis, photos_vis
CREATE TABLE IF NOT EXISTS user_settings_new (
    id INTEGER PRIMARY KEY,
    email_vis BOOLEAN NOT NULL DEFAULT 0,
    birthday_date_vis BOOLEAN NOT NULL DEFAULT 1,
    relationship_status_vis BOOLEAN NOT NULL DEFAULT 1,
    employed_at_vis BOOLEAN NOT NULL DEFAULT 1,
    phone_number_vis BOOLEAN NOT NULL DEFAULT 0,
    FOREIGN KEY (id) REFERENCES users(id) ON DELETE CASCADE
);

-- Copy existing relevant data into the new table
INSERT INTO user_settings_new (id, email_vis, birthday_date_vis, relationship_status_vis, employed_at_vis, phone_number_vis)
    SELECT id, email_vis, birthday_date_vis, relationship_status_vis, employed_at_vis, phone_number_vis FROM user_settings;

-- Drop old table and rename the new one
DROP TABLE user_settings;
ALTER TABLE user_settings_new RENAME TO user_settings;

