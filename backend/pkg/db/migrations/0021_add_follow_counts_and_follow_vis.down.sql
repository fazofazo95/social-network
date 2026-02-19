-- Rollback 0021: remove follow_vis and follow counters
ALTER TABLE users DROP COLUMN Followers;
ALTER TABLE users DROP COLUMN Following;
ALTER TABLE user_settings DROP COLUMN follow_vis;
