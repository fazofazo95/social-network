package repository

import (
	"backend/pkg/models"
	"context"
	"database/sql"
	"errors"
)

// AcceptGroupJoinRequest approves a pending join request
func (r *sqliteGroupRepo) AcceptGroupJoinRequest(ctx context.Context, approverID, groupID, requesterID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var approverRole string
	if err := tx.QueryRowContext(ctx, `
		SELECT role
		FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'active'
	`, groupID, approverID).Scan(&approverRole); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotGroupModeratorOrOwner
		}
		return err
	}

	if approverRole != "owner" && approverRole != "moderator" {
		return ErrNotGroupModeratorOrOwner
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE group_members
		SET status = 'active', role = 'member', joined_at = CURRENT_TIMESTAMP
		WHERE group_id = ? AND user_id = ? AND status = 'requested'
	`, groupID, requesterID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGroupJoinRequestNotFound
	}

	res, err = tx.ExecContext(ctx, `
		UPDATE group_join_requests
		SET status = 'approved', action_by_id = ?, request_type = 'request'
		WHERE group_id = ? AND user_id = ? AND status = 'request'
	`, approverID, groupID, requesterID)
	if err != nil {
		return err
	}

	rows, err = res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGroupJoinRequestNotFound
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE groups
		SET group_members = group_members + 1
		WHERE id = ?
	`, groupID); err != nil {
		return err
	}

	if err := r.ensureGroupChatParticipantTx(ctx, tx, groupID, requesterID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// RejectGroupJoinRequest rejects a pending join request
func (r *sqliteGroupRepo) RejectGroupJoinRequest(ctx context.Context, approverID, groupID, requesterID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var approverRole string
	if err := tx.QueryRowContext(ctx, `
		SELECT role
		FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'active'
	`, groupID, approverID).Scan(&approverRole); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotGroupModeratorOrOwner
		}
		return err
	}

	if approverRole != "owner" && approverRole != "moderator" {
		return ErrNotGroupModeratorOrOwner
	}

	res, err := tx.ExecContext(ctx, `
		DELETE FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'requested'
	`, groupID, requesterID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGroupJoinRequestNotFound
	}

	res, err = tx.ExecContext(ctx, `
		UPDATE group_join_requests
		SET status = 'rejected', action_by_id = ?, request_type = 'request'
		WHERE group_id = ? AND user_id = ? AND status = 'request'
	`, approverID, groupID, requesterID)
	if err != nil {
		return err
	}

	rows, err = res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGroupJoinRequestNotFound
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// InviteUserToGroup invites a user to a group
func (r *sqliteGroupRepo) InviteUserToGroup(ctx context.Context, inviterID, groupID, targetUserID int) (string, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var visibility, joinMode string
	if err := tx.QueryRowContext(ctx, `
		SELECT visibility, join_mode
		FROM groups
		WHERE id = ?
	`, groupID).Scan(&visibility, &joinMode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrGroupNotFound
		}
		return "", err
	}

	if visibility == "private" {
		return "", ErrPrivateGroup
	}

	var inviterRole string
	if err := tx.QueryRowContext(ctx, `
		SELECT role
		FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'active'
	`, groupID, inviterID).Scan(&inviterRole); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", ErrNotGroupModeratorOrOwner
		}
		return "", err
	}
	if inviterRole != "owner" && inviterRole != "moderator" {
		return "", ErrNotGroupModeratorOrOwner
	}

	var existingStatus string
	err = tx.QueryRowContext(ctx, `
		SELECT status
		FROM group_members
		WHERE group_id = ? AND user_id = ?
	`, groupID, targetUserID).Scan(&existingStatus)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if err == nil {
		if existingStatus == "requested" {
			return "", ErrAlreadyInvitedToGroup
		}
		return "", ErrAlreadyGroupMember
	}

	var membershipStatus string
	switch joinMode {
	case "auto", "request", "request_and_invite", "invite":
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO group_members (group_id, user_id, role, status)
			VALUES (?, ?, 'member', 'requested')
		`, groupID, targetUserID); err != nil {
			if err.Error() == "UNIQUE constraint failed: group_members.group_id, group_members.user_id" {
				return "", ErrAlreadyInvitedToGroup
			}
			return "", err
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO group_join_requests (group_id, user_id, request_type, status, action_by_id)
			VALUES (?, ?, 'invite', 'invite', ?)
		`, groupID, targetUserID, inviterID); err != nil {
			if err.Error() == "UNIQUE constraint failed: idx_group_join_requests_open_unique" {
				return "", ErrAlreadyInvitedToGroup
			}
			return "", err
		}

		membershipStatus = "requested"

	default:
		return "", errors.New("invalid group join_mode")
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return membershipStatus, nil
}

// AcceptGroupInvite accepts a group invitation
func (r *sqliteGroupRepo) AcceptGroupInvite(ctx context.Context, userID, groupID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		UPDATE group_members
		SET status = 'active', role = 'member', joined_at = CURRENT_TIMESTAMP
		WHERE group_id = ? AND user_id = ? AND status = 'requested'
	`, groupID, userID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGroupInviteNotFound
	}

	res, err = tx.ExecContext(ctx, `
		UPDATE group_join_requests
		SET status = 'accepted', action_by_id = ?, request_type = 'invite'
		WHERE group_id = ? AND user_id = ? AND status = 'invite'
	`, userID, groupID, userID)
	if err != nil {
		return err
	}

	rows, err = res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGroupInviteNotFound
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE groups
		SET group_members = group_members + 1
		WHERE id = ?
	`, groupID); err != nil {
		return err
	}

	if err := r.ensureGroupChatParticipantTx(ctx, tx, groupID, userID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// RejectGroupInvite rejects a group invitation
func (r *sqliteGroupRepo) RejectGroupInvite(ctx context.Context, userID, groupID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		DELETE FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'requested'
	`, groupID, userID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGroupInviteNotFound
	}

	res, err = tx.ExecContext(ctx, `
		UPDATE group_join_requests
		SET status = 'rejected', action_by_id = ?, request_type = 'invite'
		WHERE group_id = ? AND user_id = ? AND status = 'invite'
	`, userID, groupID, userID)
	if err != nil {
		return err
	}

	rows, err = res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGroupInviteNotFound
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// KickGroupMember removes a member from a group
func (r *sqliteGroupRepo) KickGroupMember(ctx context.Context, actorID, groupID, targetUserID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var actorRole string
	if err := tx.QueryRowContext(ctx, `
		SELECT role
		FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'active'
	`, groupID, actorID).Scan(&actorRole); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotGroupModeratorOrOwner
		}
		return err
	}
	if actorRole != "owner" && actorRole != "moderator" {
		return ErrNotGroupModeratorOrOwner
	}

	var targetRole string
	if err := tx.QueryRowContext(ctx, `
		SELECT role
		FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'active'
	`, groupID, targetUserID).Scan(&targetRole); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrGroupMemberNotFound
		}
		return err
	}

	if targetRole != "member" {
		return ErrCannotKickGroupStaff
	}

	res, err := tx.ExecContext(ctx, `
		DELETE FROM group_members
		WHERE group_id = ? AND user_id = ? AND role = 'member' AND status = 'active'
	`, groupID, targetUserID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGroupMemberNotFound
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE groups
		SET group_members = CASE WHEN group_members > 0 THEN group_members - 1 ELSE 0 END
		WHERE id = ?
	`, groupID); err != nil {
		return err
	}

	if err := r.removeGroupChatParticipantTx(ctx, tx, groupID, targetUserID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// LeaveGroup removes a user from a group
func (r *sqliteGroupRepo) LeaveGroup(ctx context.Context, userID, groupID int) (LeaveGroupResult, error) {
	result := LeaveGroupResult{}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return result, err
	}
	defer tx.Rollback()

	var role string
	if err := tx.QueryRowContext(ctx, `
		SELECT role
		FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'active'
	`, groupID, userID).Scan(&role); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return result, ErrGroupMemberNotFound
		}
		return result, err
	}

	if role != "owner" {
		res, err := tx.ExecContext(ctx, `
			DELETE FROM group_members
			WHERE group_id = ? AND user_id = ? AND status = 'active'
		`, groupID, userID)
		if err != nil {
			return result, err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return result, err
		}
		if rows == 0 {
			return result, ErrGroupMemberNotFound
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE groups
			SET group_members = CASE WHEN group_members > 0 THEN group_members - 1 ELSE 0 END
			WHERE id = ?
		`, groupID); err != nil {
			return result, err
		}

		if err := r.removeGroupChatParticipantTx(ctx, tx, groupID, userID); err != nil {
			return result, err
		}

		if err := tx.Commit(); err != nil {
			return result, err
		}

		return result, nil
	}

	newOwnerID := 0
	err = tx.QueryRowContext(ctx, `
		SELECT user_id
		FROM group_members
		WHERE group_id = ? AND status = 'active' AND role = 'moderator' AND user_id <> ?
		ORDER BY RANDOM()
		LIMIT 1
	`, groupID, userID).Scan(&newOwnerID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return result, err
	}

	if errors.Is(err, sql.ErrNoRows) {
		err = tx.QueryRowContext(ctx, `
			SELECT user_id
			FROM group_members
			WHERE group_id = ? AND status = 'active' AND role = 'member' AND user_id <> ?
			ORDER BY RANDOM()
			LIMIT 1
		`, groupID, userID).Scan(&newOwnerID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return result, err
		}
	}

	if newOwnerID == 0 {
		res, err := tx.ExecContext(ctx, `DELETE FROM groups WHERE id = ? AND owner_id = ?`, groupID, userID)
		if err != nil {
			return result, err
		}

		rows, err := res.RowsAffected()
		if err != nil {
			return result, err
		}
		if rows == 0 {
			return result, ErrGroupNotFound
		}

		if err := tx.Commit(); err != nil {
			return result, err
		}

		result.GroupDeleted = true
		return result, nil
	}

	if _, err := tx.ExecContext(ctx, `UPDATE groups SET owner_id = ? WHERE id = ?`, newOwnerID, groupID); err != nil {
		return result, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE group_members
		SET role = 'owner'
		WHERE group_id = ? AND user_id = ? AND status = 'active'
	`, groupID, newOwnerID); err != nil {
		return result, err
	}

	res, err := tx.ExecContext(ctx, `
		DELETE FROM group_members
		WHERE group_id = ? AND user_id = ? AND role = 'owner' AND status = 'active'
	`, groupID, userID)
	if err != nil {
		return result, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return result, err
	}
	if rows == 0 {
		return result, ErrGroupMemberNotFound
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE groups
		SET group_members = CASE WHEN group_members > 0 THEN group_members - 1 ELSE 0 END
		WHERE id = ?
	`, groupID); err != nil {
		return result, err
	}

	if err := r.removeGroupChatParticipantTx(ctx, tx, groupID, userID); err != nil {
		return result, err
	}

	if err := tx.Commit(); err != nil {
		return result, err
	}

	result.OwnerTransferred = true
	result.NewOwnerID = newOwnerID
	return result, nil
}

// PromoteGroupModerator promotes a member to moderator
func (r *sqliteGroupRepo) PromoteGroupModerator(ctx context.Context, ownerID, groupID, targetUserID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var actualOwnerID int
	if err := tx.QueryRowContext(ctx, `SELECT owner_id FROM groups WHERE id = ?`, groupID).Scan(&actualOwnerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrGroupNotFound
		}
		return err
	}
	if actualOwnerID != ownerID {
		return ErrNotGroupOwner
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE group_members
		SET role = 'moderator'
		WHERE group_id = ? AND user_id = ? AND status = 'active' AND role = 'member'
	`, groupID, targetUserID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		var exists int
		if err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM group_members
			WHERE group_id = ? AND user_id = ? AND status = 'active'
		`, groupID, targetUserID).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			return ErrGroupMemberNotFound
		}
		return ErrGroupMemberRoleMismatch
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// DemoteGroupModerator demotes a moderator to member
func (r *sqliteGroupRepo) DemoteGroupModerator(ctx context.Context, ownerID, groupID, targetUserID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var actualOwnerID int
	if err := tx.QueryRowContext(ctx, `SELECT owner_id FROM groups WHERE id = ?`, groupID).Scan(&actualOwnerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrGroupNotFound
		}
		return err
	}
	if actualOwnerID != ownerID {
		return ErrNotGroupOwner
	}

	res, err := tx.ExecContext(ctx, `
		UPDATE group_members
		SET role = 'member'
		WHERE group_id = ? AND user_id = ? AND status = 'active' AND role = 'moderator'
	`, groupID, targetUserID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		var exists int
		if err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM group_members
			WHERE group_id = ? AND user_id = ? AND status = 'active'
		`, groupID, targetUserID).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			return ErrGroupMemberNotFound
		}
		return ErrGroupMemberRoleMismatch
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// UpdateGroupSettings updates group visibility and join mode
func (r *sqliteGroupRepo) UpdateGroupSettings(ctx context.Context, ownerID, groupID int, visibility, joinMode *string) (GroupSettingsResult, error) {
	out := GroupSettingsResult{GroupID: groupID}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return out, err
	}
	defer tx.Rollback()

	var actualOwnerID int
	var currentVisibility, currentJoinMode string
	if err := tx.QueryRowContext(ctx, `
		SELECT owner_id, visibility, join_mode
		FROM groups
		WHERE id = ?
	`, groupID).Scan(&actualOwnerID, &currentVisibility, &currentJoinMode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return out, ErrGroupNotFound
		}
		return out, err
	}

	if actualOwnerID != ownerID {
		return out, ErrNotGroupOwner
	}

	newVisibility := currentVisibility
	if visibility != nil {
		newVisibility = *visibility
	}

	newJoinMode := currentJoinMode
	if joinMode != nil {
		newJoinMode = *joinMode
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE groups
		SET visibility = ?, join_mode = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, newVisibility, newJoinMode, groupID); err != nil {
		return out, err
	}

	if err := tx.Commit(); err != nil {
		return out, err
	}

	out.Visibility = newVisibility
	out.JoinMode = newJoinMode
	return out, nil
}

// GetGroupSettings retrieves group settings
func (r *sqliteGroupRepo) GetGroupSettings(ctx context.Context, ownerID, groupID int) (GroupSettingsResult, error) {
	out := GroupSettingsResult{GroupID: groupID}

	var actualOwnerID int
	if err := r.db.QueryRowContext(ctx, `
		SELECT owner_id, visibility, join_mode
		FROM groups
		WHERE id = ?
	`, groupID).Scan(&actualOwnerID, &out.Visibility, &out.JoinMode); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return out, ErrGroupNotFound
		}
		return out, err
	}

	if actualOwnerID != ownerID {
		return out, ErrNotGroupOwner
	}

	return out, nil
}

// GetActiveGroupMembers retrieves active members of a group
func (r *sqliteGroupRepo) GetActiveGroupMembers(ctx context.Context, groupID int) ([]models.GroupMemberListItem, error) {
	var groupExists int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM groups WHERE id = ?`, groupID).Scan(&groupExists); err != nil {
		return nil, err
	}
	if groupExists == 0 {
		return nil, ErrGroupNotFound
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.first_name, u.last_name, COALESCE(u.profile_picture, ''), gm.role
		FROM group_members gm
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id = ? AND gm.status = 'active'
		ORDER BY CASE gm.role WHEN 'owner' THEN 0 WHEN 'moderator' THEN 1 ELSE 2 END,
			u.first_name, u.last_name
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.GroupMemberListItem, 0)
	for rows.Next() {
		var item models.GroupMemberListItem
		if err := rows.Scan(&item.ID, &item.FirstName, &item.LastName, &item.ProfilePicture, &item.Role); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	return out, nil
}

// GetPendingGroupJoinRequests retrieves pending join requests
func (r *sqliteGroupRepo) GetPendingGroupJoinRequests(ctx context.Context, actorID, groupID int) ([]models.GroupPendingItem, error) {
	if err := r.isGroupModeratorOrOwnerCheck(ctx, groupID, actorID); err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.first_name, u.last_name, COALESCE(u.profile_picture, ''), 'requested' AS type
		FROM group_join_requests gjr
		JOIN users u ON u.id = gjr.user_id
		WHERE gjr.group_id = ? AND gjr.request_type = 'request' AND gjr.status = 'request'
		ORDER BY gjr.created_at DESC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.GroupPendingItem, 0)
	for rows.Next() {
		var item models.GroupPendingItem
		if err := rows.Scan(&item.ID, &item.FirstName, &item.LastName, &item.ProfilePicture, &item.Type); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	return out, nil
}

// GetPendingGroupInvites retrieves pending invites
func (r *sqliteGroupRepo) GetPendingGroupInvites(ctx context.Context, actorID, groupID int) ([]models.GroupPendingItem, error) {
	if err := r.isGroupModeratorOrOwnerCheck(ctx, groupID, actorID); err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.first_name, u.last_name, COALESCE(u.profile_picture, ''), 'invited' AS type
		FROM group_join_requests gjr
		JOIN users u ON u.id = gjr.user_id
		WHERE gjr.group_id = ? AND gjr.request_type = 'invite' AND gjr.status = 'invite'
		ORDER BY gjr.created_at DESC
	`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.GroupPendingItem, 0)
	for rows.Next() {
		var item models.GroupPendingItem
		if err := rows.Scan(&item.ID, &item.FirstName, &item.LastName, &item.ProfilePicture, &item.Type); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	return out, nil
}

// GetGroupPageView retrieves the group page view details
func (r *sqliteGroupRepo) GetGroupPageView(ctx context.Context, viewerID, groupID int) (GroupPageView, error) {
	var out GroupPageView

	if err := r.db.QueryRowContext(ctx, `
		SELECT id, name, COALESCE(description, ''), visibility, join_mode, COALESCE(group_picture, ''),
			group_members, COALESCE(datetime(created_at), '')
		FROM groups
		WHERE id = ?
	`, groupID).Scan(
		&out.ID,
		&out.Name,
		&out.Description,
		&out.Visibility,
		&out.JoinMode,
		&out.GroupPicture,
		&out.GroupMembers,
		&out.CreatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return out, ErrGroupNotFound
		}
		return out, err
	}

	var role string
	err := r.db.QueryRowContext(ctx, `
		SELECT role
		FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'active'
	`, groupID, viewerID).Scan(&role)
	if err == nil {
		out.IsActive = true
		out.Role = role
		return out, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return out, err
	}

	var requestType string
	err = r.db.QueryRowContext(ctx, `
		SELECT request_type
		FROM group_join_requests
		WHERE group_id = ? AND user_id = ? AND status IN ('request', 'invite')
		ORDER BY CASE status WHEN 'invite' THEN 0 ELSE 1 END, created_at DESC
		LIMIT 1
	`, groupID, viewerID).Scan(&requestType)
	if err == nil {
		switch requestType {
		case "invite":
			out.PendingType = "invited"
		case "request":
			out.PendingType = "requested"
		}
		return out, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return out, err
	}

	out.PendingType = "none"
	return out, nil
}

// DiscoverGroups discovers public groups
func (r *sqliteGroupRepo) DiscoverGroups(ctx context.Context, userID, limit, offset int) ([]models.GroupDiscoverItem, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT g.id, g.name, COALESCE(g.description, ''), COALESCE(g.group_picture, ''), g.group_members,
		       COALESCE(u.first_name, ''), COALESCE(u.last_name, ''), g.join_mode
		FROM groups g
		JOIN users u ON u.id = g.owner_id
		WHERE g.visibility = 'public'
		  AND g.join_mode <> 'invite'
		  AND NOT EXISTS (
			SELECT 1
			FROM group_members gm
			WHERE gm.group_id = g.id
			  AND gm.user_id = ?
			  AND gm.status IN ('active', 'requested', 'invited')
		  )
		  AND NOT EXISTS (
			SELECT 1
			FROM group_join_requests gjr
			WHERE gjr.group_id = g.id
			  AND gjr.user_id = ?
			  AND gjr.status IN ('request', 'invite')
		  )
		ORDER BY g.created_at DESC, g.id DESC
		LIMIT ? OFFSET ?
	`, userID, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.GroupDiscoverItem, 0)
	for rows.Next() {
		var item models.GroupDiscoverItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Description, &item.GroupPicture, &item.GroupMembers, &item.OwnerFirst, &item.OwnerLast, &item.Type); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

// GetActiveGroupsForUser retrieves active groups for a user
func (r *sqliteGroupRepo) GetActiveGroupsForUser(ctx context.Context, userID int) ([]models.GroupActiveItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT g.id, g.name, COALESCE(g.description, ''), COALESCE(g.group_picture, ''), g.group_members,
		       g.owner_id, COALESCE(u.first_name, ''), COALESCE(u.last_name, ''), gm.role,
		       COALESCE(datetime(g.created_at), '')
		FROM group_members gm
		JOIN groups g ON g.id = gm.group_id
		JOIN users u ON u.id = g.owner_id
		WHERE gm.user_id = ? AND gm.status = 'active'
		ORDER BY g.created_at DESC, g.id DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.GroupActiveItem, 0)
	for rows.Next() {
		var item models.GroupActiveItem
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.GroupPicture,
			&item.GroupMembers,
			&item.OwnerID,
			&item.OwnerFirst,
			&item.OwnerLast,
			&item.Role,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

// GetUserPendingGroupRequests retrieves pending requests where user is the requester
func (r *sqliteGroupRepo) GetUserPendingGroupRequests(ctx context.Context, userID int) ([]models.GroupUserPendingItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT g.id, g.name, COALESCE(g.description, ''), COALESCE(g.group_picture, ''), g.group_members,
		       g.owner_id, COALESCE(u.first_name, ''), COALESCE(u.last_name, ''), g.join_mode,
		       COALESCE(datetime(gjr.created_at), ''), 'requested' AS type
		FROM group_join_requests gjr
		JOIN groups g ON g.id = gjr.group_id
		JOIN users u ON u.id = g.owner_id
		WHERE gjr.user_id = ? AND gjr.request_type = 'request' AND gjr.status = 'request'
		ORDER BY gjr.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.GroupUserPendingItem, 0)
	for rows.Next() {
		var item models.GroupUserPendingItem
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.GroupPicture,
			&item.GroupMembers,
			&item.OwnerID,
			&item.OwnerFirst,
			&item.OwnerLast,
			&item.JoinMode,
			&item.RequestedAt,
			&item.Type,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

// GetUserPendingGroupInvites retrieves pending invites for a user
func (r *sqliteGroupRepo) GetUserPendingGroupInvites(ctx context.Context, userID int) ([]models.GroupUserPendingItem, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT g.id, g.name, COALESCE(g.description, ''), COALESCE(g.group_picture, ''), g.group_members,
		       g.owner_id, COALESCE(u.first_name, ''), COALESCE(u.last_name, ''), g.join_mode,
		       COALESCE(datetime(gjr.created_at), ''), 'invited' AS type
		FROM group_join_requests gjr
		JOIN groups g ON g.id = gjr.group_id
		JOIN users u ON u.id = g.owner_id
		WHERE gjr.user_id = ? AND gjr.request_type = 'invite' AND gjr.status = 'invite'
		ORDER BY gjr.created_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.GroupUserPendingItem, 0)
	for rows.Next() {
		var item models.GroupUserPendingItem
		if err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Description,
			&item.GroupPicture,
			&item.GroupMembers,
			&item.OwnerID,
			&item.OwnerFirst,
			&item.OwnerLast,
			&item.JoinMode,
			&item.RequestedAt,
			&item.Type,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

// RemovePendingGroupInvite removes an invite sent to a user
func (r *sqliteGroupRepo) RemovePendingGroupInvite(ctx context.Context, actorID, groupID, targetUserID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var actorRole string
	if err := tx.QueryRowContext(ctx, `
		SELECT role
		FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'active'
	`, groupID, actorID).Scan(&actorRole); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			var groupExists int
			if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM groups WHERE id = ?`, groupID).Scan(&groupExists); err != nil {
				return err
			}
			if groupExists == 0 {
				return ErrGroupNotFound
			}
			return ErrNotGroupModeratorOrOwner
		}
		return err
	}
	if actorRole != "owner" && actorRole != "moderator" {
		return ErrNotGroupModeratorOrOwner
	}

	var memberStatus string
	err = tx.QueryRowContext(ctx, `
		SELECT status
		FROM group_members
		WHERE group_id = ? AND user_id = ?
	`, groupID, targetUserID).Scan(&memberStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrGroupInviteNotFound
		}
		return err
	}
	if memberStatus == "active" {
		return ErrGroupMemberIsActive
	}

	res, err := tx.ExecContext(ctx, `
		DELETE FROM group_join_requests
		WHERE group_id = ? AND user_id = ? AND request_type = 'invite' AND status = 'invite'
	`, groupID, targetUserID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGroupInviteNotFound
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM group_members
		WHERE group_id = ? AND user_id = ? AND status <> 'active'
	`, groupID, targetUserID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// RemoveOwnPendingGroupRequest removes own pending join request
func (r *sqliteGroupRepo) RemoveOwnPendingGroupRequest(ctx context.Context, userID, groupID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var memberStatus string
	err = tx.QueryRowContext(ctx, `
		SELECT status
		FROM group_members
		WHERE group_id = ? AND user_id = ?
	`, groupID, userID).Scan(&memberStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrGroupJoinRequestNotFound
		}
		return err
	}
	if memberStatus == "active" {
		return ErrGroupMemberIsActive
	}

	res, err := tx.ExecContext(ctx, `
		DELETE FROM group_join_requests
		WHERE group_id = ? AND user_id = ? AND request_type = 'request' AND status = 'request'
	`, groupID, userID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGroupJoinRequestNotFound
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM group_members
		WHERE group_id = ? AND user_id = ? AND status <> 'active'
	`, groupID, userID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// CreateGroupEvent creates a new group event
func (r *sqliteGroupRepo) CreateGroupEvent(ctx context.Context, actorID, groupID int, in models.GroupEventCreateInput) (*models.GroupEvent, error) {
	if err := r.isGroupModeratorOrOwnerCheck(ctx, groupID, actorID); err != nil {
		return nil, err
	}

	var out models.GroupEvent

	res, err := r.db.ExecContext(ctx, `
		INSERT INTO group_events (group_id, creator_id, title, description, event_day, event_time, going)
		VALUES (?, ?, ?, ?, ?, ?, 1)
	`, groupID, actorID, in.Title, in.Description, in.EventDay, in.EventTime)
	if err != nil {
		return nil, err
	}

	newID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO group_events_reaction (event_id, group_id, creator_id, reactor_id, reaction_type)
		VALUES (?, ?, ?, ?, 'going')
	`, newID, groupID, actorID, actorID); err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx, `
		SELECT id, group_id, creator_id, title, COALESCE(description, ''),
		       COALESCE(date(event_day), ''), COALESCE(time(event_time), ''),
		       COALESCE(datetime(created_at), ''), going, not_going, invited
		FROM group_events
		WHERE id = ?
	`, newID).Scan(
		&out.ID,
		&out.GroupID,
		&out.CreatorID,
		&out.Title,
		&out.Description,
		&out.EventDay,
		&out.EventTime,
		&out.CreatedAt,
		&out.Going,
		&out.NotGoing,
		&out.Invited,
	); err != nil {
		return nil, err
	}

	return &out, nil
}

// GetGroupEventInviteableMembers retrieves members who can be invited to an event
func (r *sqliteGroupRepo) GetGroupEventInviteableMembers(ctx context.Context, actorID, groupID, eventID int) ([]models.GroupMemberListItem, error) {
	if err := r.isActiveGroupMemberCheck(ctx, groupID, actorID); err != nil {
		return nil, err
	}

	if err := r.hasGroupEventReactionCheck(ctx, eventID, actorID); err != nil {
		return nil, err
	}

	if err := r.getGroupEventCreatorCheck(ctx, groupID, eventID); err != nil {
		return nil, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT u.id, u.first_name, u.last_name, COALESCE(u.profile_picture, ''), gm.role
		FROM group_members gm
		JOIN users u ON u.id = gm.user_id
		WHERE gm.group_id = ? AND gm.status = 'active'
		  AND gm.user_id <> ?
		  AND NOT EXISTS (
			SELECT 1
			FROM group_events_reaction ger
			WHERE ger.event_id = ? AND ger.reactor_id = gm.user_id
		  )
		ORDER BY CASE gm.role WHEN 'owner' THEN 0 WHEN 'moderator' THEN 1 ELSE 2 END,
			u.first_name, u.last_name
	`, groupID, actorID, eventID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.GroupMemberListItem, 0)
	for rows.Next() {
		var item models.GroupMemberListItem
		if err := rows.Scan(&item.ID, &item.FirstName, &item.LastName, &item.ProfilePicture, &item.Role); err != nil {
			return nil, err
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return out, nil
}

// InviteGroupEventMember invites a member to an event
func (r *sqliteGroupRepo) InviteGroupEventMember(ctx context.Context, actorID, groupID, eventID, targetUserID int) error {
	if actorID == targetUserID {
		return ErrCannotInviteSelf
	}

	if err := r.isActiveGroupMemberCheck(ctx, groupID, actorID); err != nil {
		return err
	}

	if err := r.hasGroupEventReactionCheck(ctx, eventID, actorID); err != nil {
		return err
	}

	if err := r.getGroupEventCreatorCheck(ctx, groupID, eventID); err != nil {
		return err
	}

	if err := r.isActiveGroupMemberCheck(ctx, groupID, targetUserID); err != nil {
		return ErrTargetNotActiveMember
	}

	var existing int
	if err := r.db.QueryRowContext(ctx, `
		SELECT 1
		FROM group_events_reaction
		WHERE event_id = ? AND reactor_id = ?
		LIMIT 1
	`, eventID, targetUserID).Scan(&existing); err == nil {
		return ErrGroupEventAlreadyAnswered
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	creatorID, err := r.getGroupEventCreatorID(ctx, groupID, eventID)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO group_events_reaction (event_id, group_id, creator_id, reactor_id, reaction_type)
		VALUES (?, ?, ?, ?, 'invited')
	`, eventID, groupID, creatorID, targetUserID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE group_events
		SET invited = invited + 1
		WHERE id = ?
	`, eventID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// InviteAllGroupEventMembers invites all eligible members to an event
func (r *sqliteGroupRepo) InviteAllGroupEventMembers(ctx context.Context, actorID, groupID, eventID int) (int, error) {
	if err := r.isActiveGroupMemberCheck(ctx, groupID, actorID); err != nil {
		return 0, err
	}

	if err := r.hasGroupEventReactionCheck(ctx, eventID, actorID); err != nil {
		return 0, err
	}

	if err := r.getGroupEventCreatorCheck(ctx, groupID, eventID); err != nil {
		return 0, err
	}

	creatorID, err := r.getGroupEventCreatorID(ctx, groupID, eventID)
	if err != nil {
		return 0, err
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO group_events_reaction (event_id, group_id, creator_id, reactor_id, reaction_type)
		SELECT ?, ?, ?, gm.user_id, 'invited'
		FROM group_members gm
		WHERE gm.group_id = ? AND gm.status = 'active'
		  AND gm.user_id <> ?
		  AND NOT EXISTS (
			SELECT 1
			FROM group_events_reaction ger
			WHERE ger.event_id = ? AND ger.reactor_id = gm.user_id
		  )
	`, eventID, groupID, creatorID, groupID, actorID, eventID)
	if err != nil {
		return 0, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	if rows > 0 {
		if _, err := tx.ExecContext(ctx, `
			UPDATE group_events
			SET invited = invited + ?
			WHERE id = ?
		`, rows, eventID); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return int(rows), nil
}

// RespondToGroupEventInvite responds to an event invitation
func (r *sqliteGroupRepo) RespondToGroupEventInvite(ctx context.Context, actorID, groupID, eventID int, reactionType string) error {
	if err := r.isActiveGroupMemberCheck(ctx, groupID, actorID); err != nil {
		return err
	}

	if reactionType != "going" && reactionType != "not_going" {
		return errors.New("invalid reaction type")
	}

	if err := r.getGroupEventCreatorCheck(ctx, groupID, eventID); err != nil {
		return err
	}

	var current string
	err := r.db.QueryRowContext(ctx, `
		SELECT reaction_type
		FROM group_events_reaction
		WHERE event_id = ? AND reactor_id = ?
	`, eventID, actorID).Scan(&current)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotInvitedToEvent
		}
		return err
	}

	if current != "invited" {
		return ErrGroupEventAlreadyResponded
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE group_events_reaction
		SET reaction_type = ?
		WHERE event_id = ? AND reactor_id = ?
	`, reactionType, eventID, actorID); err != nil {
		return err
	}

	if reactionType == "going" {
		if _, err := tx.ExecContext(ctx, `
			UPDATE group_events
			SET invited = CASE WHEN invited > 0 THEN invited - 1 ELSE 0 END,
			    going = going + 1
			WHERE id = ?
		`, eventID); err != nil {
			return err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `
			UPDATE group_events
			SET invited = CASE WHEN invited > 0 THEN invited - 1 ELSE 0 END,
			    not_going = not_going + 1
			WHERE id = ?
		`, eventID); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// ChangeGroupEventResponse changes an existing event response
func (r *sqliteGroupRepo) ChangeGroupEventResponse(ctx context.Context, actorID, groupID, eventID int, reactionType string) error {
	if err := r.isActiveGroupMemberCheck(ctx, groupID, actorID); err != nil {
		return err
	}

	if reactionType != "going" && reactionType != "not_going" {
		return errors.New("invalid reaction type")
	}

	if err := r.getGroupEventCreatorCheck(ctx, groupID, eventID); err != nil {
		return err
	}

	var current string
	err := r.db.QueryRowContext(ctx, `
		SELECT reaction_type
		FROM group_events_reaction
		WHERE event_id = ? AND reactor_id = ?
	`, eventID, actorID).Scan(&current)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrGroupEventNoResponseToChange
		}
		return err
	}

	if current != "going" && current != "not_going" {
		return ErrGroupEventNoResponseToChange
	}
	if current == reactionType {
		return ErrGroupEventResponseUnchanged
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		UPDATE group_events_reaction
		SET reaction_type = ?
		WHERE event_id = ? AND reactor_id = ?
	`, reactionType, eventID, actorID); err != nil {
		return err
	}

	if current == "going" {
		if _, err := tx.ExecContext(ctx, `
			UPDATE group_events
			SET going = CASE WHEN going > 0 THEN going - 1 ELSE 0 END,
			    not_going = not_going + 1
			WHERE id = ?
		`, eventID); err != nil {
			return err
		}
	} else {
		if _, err := tx.ExecContext(ctx, `
			UPDATE group_events
			SET not_going = CASE WHEN not_going > 0 THEN not_going - 1 ELSE 0 END,
			    going = going + 1
			WHERE id = ?
		`, eventID); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// DeleteGroupEvent deletes a group event
func (r *sqliteGroupRepo) DeleteGroupEvent(ctx context.Context, actorID, groupID, eventID int) error {
	if err := r.isGroupModeratorOrOwnerCheck(ctx, groupID, actorID); err != nil {
		return err
	}

	res, err := r.db.ExecContext(ctx, `
		DELETE FROM group_events
		WHERE id = ? AND group_id = ?
	`, eventID, groupID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGroupEventNotFound
	}

	return nil
}

// Helper functions
func (r *sqliteGroupRepo) isGroupModeratorOrOwnerCheck(ctx context.Context, groupID, userID int) error {
	var role string
	err := r.db.QueryRowContext(ctx, `
		SELECT role
		FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'active'
	`, groupID, userID).Scan(&role)
	if err == nil {
		if role == "owner" || role == "moderator" {
			return nil
		}
		return ErrNotGroupModeratorOrOwner
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	var groupExists int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM groups WHERE id = ?`, groupID).Scan(&groupExists); err != nil {
		return err
	}
	if groupExists == 0 {
		return ErrGroupNotFound
	}

	return ErrNotGroupModeratorOrOwner
}

func (r *sqliteGroupRepo) isActiveGroupMemberCheck(ctx context.Context, groupID, userID int) error {
	var exists int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1
		FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'active'
		LIMIT 1
	`, groupID, userID).Scan(&exists)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	var groupExists int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM groups WHERE id = ?`, groupID).Scan(&groupExists); err != nil {
		return err
	}
	if groupExists == 0 {
		return ErrGroupNotFound
	}

	return ErrNotActiveGroupMember
}

func (r *sqliteGroupRepo) getGroupEventCreatorCheck(ctx context.Context, groupID, eventID int) error {
	_, err := r.getGroupEventCreatorID(ctx, groupID, eventID)
	return err
}

func (r *sqliteGroupRepo) getGroupEventCreatorID(ctx context.Context, groupID, eventID int) (int, error) {
	var creatorID int
	err := r.db.QueryRowContext(ctx, `
		SELECT creator_id
		FROM group_events
		WHERE id = ? AND group_id = ?
	`, eventID, groupID).Scan(&creatorID)
	if err == nil {
		return creatorID, nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrGroupEventNotFound
	}
	return 0, err
}

func (r *sqliteGroupRepo) hasGroupEventReactionCheck(ctx context.Context, eventID, userID int) error {
	var exists int
	err := r.db.QueryRowContext(ctx, `
		SELECT 1
		FROM group_events_reaction
		WHERE event_id = ? AND reactor_id = ?
		LIMIT 1
	`, eventID, userID).Scan(&exists)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	return ErrNotInvitedToEvent
}
