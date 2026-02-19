-- Rollback for 0020: remove about_me_vis and nickname_vis by recreating table

-- Recreate table without the two new columns
CREATE TABLE IF NOT EXISTS user_settings_old (
    id INTEGER PRIMARY KEY,
    email_vis BOOLEAN NOT NULL DEFAULT 0,
    birthday_date_vis BOOLEAN NOT NULL DEFAULT 1,
    relationship_status_vis BOOLEAN NOT NULL DEFAULT 1,
    employed_at_vis BOOLEAN NOT NULL DEFAULT 1,
    phone_number_vis BOOLEAN NOT NULL DEFAULT 0,
    FOREIGN KEY (id) REFERENCES users(id) ON DELETE CASCADE
);

INSERT INTO user_settings_old (id, email_vis, birthday_date_vis, relationship_status_vis, employed_at_vis, phone_number_vis)
    SELECT id, email_vis, birthday_date_vis, relationship_status_vis, employed_at_vis, phone_number_vis FROM user_settings;

DROP TABLE user_settings;
ALTER TABLE user_settings_old RENAME TO user_settings;
