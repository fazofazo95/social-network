-- Rollback for 0019: add back the removed columns to user_settings

-- Recreate the original table including posts_vis, created_date_vis, photos_vis
CREATE TABLE IF NOT EXISTS user_settings_old (
    id INTEGER PRIMARY KEY,
    posts_vis BOOLEAN NOT NULL DEFAULT 1,
    email_vis BOOLEAN NOT NULL DEFAULT 0,
    created_date_vis BOOLEAN NOT NULL DEFAULT 0,
    birthday_date_vis BOOLEAN NOT NULL DEFAULT 1,
    relationship_status_vis BOOLEAN NOT NULL DEFAULT 1,
    employed_at_vis BOOLEAN NOT NULL DEFAULT 1,
    phone_number_vis BOOLEAN NOT NULL DEFAULT 0,
    photos_vis BOOLEAN NOT NULL DEFAULT 1,
    FOREIGN KEY (id) REFERENCES users(id) ON DELETE CASCADE
);

-- Copy existing data back and set reasonable defaults for the restored columns
INSERT INTO user_settings_old (id, posts_vis, email_vis, created_date_vis, birthday_date_vis, relationship_status_vis, employed_at_vis, phone_number_vis, photos_vis)
    SELECT id,
           1 AS posts_vis,
           email_vis,
           0 AS created_date_vis,
           birthday_date_vis,
           relationship_status_vis,
           employed_at_vis,
           phone_number_vis,
           1 AS photos_vis
    FROM user_settings;

DROP TABLE user_settings;
ALTER TABLE user_settings_old RENAME TO user_settings;

