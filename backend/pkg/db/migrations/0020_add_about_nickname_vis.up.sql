-- Migration 0020: add about_me_vis and nickname_vis to user_settings

ALTER TABLE user_settings ADD COLUMN about_me_vis BOOLEAN NOT NULL DEFAULT 1;
ALTER TABLE user_settings ADD COLUMN nickname_vis BOOLEAN NOT NULL DEFAULT 1;
