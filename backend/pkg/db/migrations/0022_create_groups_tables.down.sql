-- Rollback 0022: drop groups-related tables

DROP INDEX IF EXISTS idx_group_join_requests_open_unique;
DROP INDEX IF EXISTS idx_group_join_requests_user_status;
DROP INDEX IF EXISTS idx_group_join_requests_group_status;
DROP TABLE IF EXISTS group_join_requests;

DROP INDEX IF EXISTS idx_group_members_status;
DROP INDEX IF EXISTS idx_group_members_user_id;
DROP TABLE IF EXISTS group_members;

DROP INDEX IF EXISTS idx_groups_created_at;
DROP INDEX IF EXISTS idx_groups_visibility;
DROP INDEX IF EXISTS idx_groups_owner_id;
DROP TABLE IF EXISTS groups;
