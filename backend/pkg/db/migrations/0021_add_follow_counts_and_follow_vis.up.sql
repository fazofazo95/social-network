-- Migration 0021: add profile follow counts and follow visibility setting
ALTER TABLE users ADD COLUMN Followers INTEGER NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN Following INTEGER NOT NULL DEFAULT 0;
ALTER TABLE user_settings ADD COLUMN follow_vis BOOLEAN NOT NULL DEFAULT 0;

-- Backfill counters from existing accepted relationships
UPDATE users
SET Followers = (SELECT COUNT(*) FROM followers WHERE followed_id = users.id AND status = 'accepted'),
	Following = (SELECT COUNT(*) FROM followers WHERE follower_id = users.id AND status = 'accepted');
