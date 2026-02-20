package queries

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"backend/pkg/models"
)

var ErrGroupNameTaken = errors.New("group name already in use")
var ErrGroupNotFound = errors.New("group not found")
var ErrNotGroupOwner = errors.New("only the group owner can delete the group")
var ErrPrivateGroup = errors.New("cannot join private group from this endpoint")
var ErrInviteOnlyGroup = errors.New("group is invite-only")
var ErrAlreadyGroupMember = errors.New("user is already a group member")
var ErrAlreadyRequestedToJoin = errors.New("user has already requested to join")
var ErrAlreadyInvitedToGroup = errors.New("user has already been invited")
var ErrNotGroupModeratorOrOwner = errors.New("only group owner or moderators can approve requests")
var ErrGroupJoinRequestNotFound = errors.New("group join request not found")
var ErrGroupInviteNotFound = errors.New("group invite not found")
var ErrGroupMemberNotFound = errors.New("group member not found")
var ErrCannotKickGroupStaff = errors.New("cannot kick owner or moderator")
var ErrGroupMemberRoleMismatch = errors.New("group member role mismatch")
var ErrGroupMemberIsActive = errors.New("group member is active")

func ensureGroupChatIDTx(ctx context.Context, tx *sql.Tx, groupID int) (int, error) {
	var chatID int
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM chats
		WHERE type = 'group' AND group_id = ?
		LIMIT 1
	`, groupID).Scan(&chatID)
	if err == nil {
		return chatID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	var ownerID int
	if err := tx.QueryRowContext(ctx, `SELECT owner_id FROM groups WHERE id = ?`, groupID).Scan(&ownerID); err != nil {
		return 0, err
	}

	res, err := tx.ExecContext(ctx, `
		INSERT INTO chats (type, group_id, created_by, last_message_at)
		VALUES ('group', ?, ?, CURRENT_TIMESTAMP)
	`, groupID, ownerID)
	if err != nil {
		return 0, err
	}

	chatID64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(chatID64), nil
}

func ensureGroupChatParticipantTx(ctx context.Context, tx *sql.Tx, groupID, userID int) error {
	chatID, err := ensureGroupChatIDTx(ctx, tx, groupID)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT OR IGNORE INTO chat_participants (chat_id, user_id)
		VALUES (?, ?)
	`, chatID, userID)
	if err != nil {
		return err
	}

	return nil
}

func removeGroupChatParticipantTx(ctx context.Context, tx *sql.Tx, groupID, userID int) error {
	_, err := tx.ExecContext(ctx, `
		DELETE FROM chat_participants
		WHERE user_id = ?
		  AND chat_id = (
			SELECT id
			FROM chats
			WHERE type = 'group' AND group_id = ?
			LIMIT 1
		  )
	`, userID, groupID)
	if err != nil {
		return err
	}

	return nil
}

func CreateGroup(ctx context.Context, db *sql.DB, ownerID int, in models.CreateGroupInput) (models.GroupResponse, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return models.GroupResponse{}, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO groups (name, description, owner_id, visibility, join_mode, group_members)
		VALUES (?, ?, ?, ?, ?, 1)
	`, in.Name, in.Description, ownerID, in.Visibility, in.JoinMode)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "groups.name") {
			return models.GroupResponse{}, ErrGroupNameTaken
		}
		return models.GroupResponse{}, err
	}

	groupID64, err := res.LastInsertId()
	if err != nil {
		return models.GroupResponse{}, err
	}
	groupID := int(groupID64)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES (?, ?, 'owner', 'active')
	`, groupID, ownerID)
	if err != nil {
		return models.GroupResponse{}, err
	}

	chatRes, err := tx.ExecContext(ctx, `
		INSERT INTO chats (type, group_id, created_by, last_message_at)
		VALUES ('group', ?, ?, CURRENT_TIMESTAMP)
	`, groupID, ownerID)
	if err != nil {
		return models.GroupResponse{}, err
	}

	chatID64, err := chatRes.LastInsertId()
	if err != nil {
		return models.GroupResponse{}, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO chat_participants (chat_id, user_id)
		VALUES (?, ?)
	`, int(chatID64), ownerID)
	if err != nil {
		return models.GroupResponse{}, err
	}

	var out models.GroupResponse
	out.ID = groupID
	out.Name = in.Name
	out.Description = in.Description
	out.OwnerID = ownerID
	out.Visibility = in.Visibility
	out.JoinMode = in.JoinMode
	out.GroupMembers = 1

	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(datetime(created_at), '') FROM groups WHERE id = ?`, groupID).Scan(&out.CreatedAt); err != nil {
		return models.GroupResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		return models.GroupResponse{}, err
	}

	return out, nil
}

func DeleteGroup(ctx context.Context, db *sql.DB, requesterID, groupID int) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var ownerID int
	if err := tx.QueryRowContext(ctx, `SELECT owner_id FROM groups WHERE id = ?`, groupID).Scan(&ownerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrGroupNotFound
		}
		return err
	}

	if ownerID != requesterID {
		return ErrNotGroupOwner
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM group_join_requests WHERE group_id = ?`, groupID); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM group_members WHERE group_id = ?`, groupID); err != nil {
		return err
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM groups WHERE id = ?`, groupID)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrGroupNotFound
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func RequestToJoinGroup(ctx context.Context, db *sql.DB, userID, groupID int) (string, error) {
	tx, err := db.BeginTx(ctx, nil)
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

	var existingStatus string
	err = tx.QueryRowContext(ctx, `
		SELECT status
		FROM group_members
		WHERE group_id = ? AND user_id = ?
	`, groupID, userID).Scan(&existingStatus)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}
	if err == nil {
		if existingStatus == "requested" {
			return "", ErrAlreadyRequestedToJoin
		}
		return "", ErrAlreadyGroupMember
	}

	var membershipStatus string
	switch joinMode {
	case "auto":
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO group_members (group_id, user_id, role, status)
			VALUES (?, ?, 'member', 'active')
		`, groupID, userID); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "group_members.group_id") {
				return "", ErrAlreadyGroupMember
			}
			return "", err
		}

		if _, err := tx.ExecContext(ctx, `
			UPDATE groups
			SET group_members = group_members + 1
			WHERE id = ?
		`, groupID); err != nil {
			return "", err
		}

		if err := ensureGroupChatParticipantTx(ctx, tx, groupID, userID); err != nil {
			return "", err
		}

		membershipStatus = "active"

	case "request", "request_and_invite":
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO group_members (group_id, user_id, role, status)
			VALUES (?, ?, 'member', 'requested')
		`, groupID, userID); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "group_members.group_id") {
				return "", ErrAlreadyRequestedToJoin
			}
			return "", err
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO group_join_requests (group_id, user_id, request_type, status)
			VALUES (?, ?, 'request', 'request')
		`, groupID, userID); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "idx_group_join_requests_open_unique") {
				return "", ErrAlreadyRequestedToJoin
			}
			return "", err
		}

		membershipStatus = "requested"

	case "invite":
		return "", ErrInviteOnlyGroup

	default:
		return "", errors.New("invalid group join_mode")
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}

	return membershipStatus, nil
}

func AcceptGroupJoinRequest(ctx context.Context, db *sql.DB, approverID, groupID, requesterID int) error {
	tx, err := db.BeginTx(ctx, nil)
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

	if err := ensureGroupChatParticipantTx(ctx, tx, groupID, requesterID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func RejectGroupJoinRequest(ctx context.Context, db *sql.DB, approverID, groupID, requesterID int) error {
	tx, err := db.BeginTx(ctx, nil)
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

func InviteUserToGroup(ctx context.Context, db *sql.DB, inviterID, groupID, targetUserID int) (string, error) {
	tx, err := db.BeginTx(ctx, nil)
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
			if strings.Contains(strings.ToLower(err.Error()), "group_members.group_id") {
				return "", ErrAlreadyInvitedToGroup
			}
			return "", err
		}

		if _, err := tx.ExecContext(ctx, `
			INSERT INTO group_join_requests (group_id, user_id, request_type, status, action_by_id)
			VALUES (?, ?, 'invite', 'invite', ?)
		`, groupID, targetUserID, inviterID); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "idx_group_join_requests_open_unique") {
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

func AcceptGroupInvite(ctx context.Context, db *sql.DB, userID, groupID int) error {
	tx, err := db.BeginTx(ctx, nil)
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

	if err := ensureGroupChatParticipantTx(ctx, tx, groupID, userID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func RejectGroupInvite(ctx context.Context, db *sql.DB, userID, groupID int) error {
	tx, err := db.BeginTx(ctx, nil)
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

type LeaveGroupResult struct {
	GroupDeleted     bool
	OwnerTransferred bool
	NewOwnerID       int
}

type GroupSettingsResult struct {
	GroupID    int
	Visibility string
	JoinMode   string
}

type GroupPageView struct {
	ID           int
	Name         string
	Description  string
	Visibility   string
	JoinMode     string
	GroupPicture string
	GroupMembers int
	CreatedAt    string
	IsActive     bool
	Role         string
	PendingType  string
}

func KickGroupMember(ctx context.Context, db *sql.DB, actorID, groupID, targetUserID int) error {
	tx, err := db.BeginTx(ctx, nil)
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

	if err := removeGroupChatParticipantTx(ctx, tx, groupID, targetUserID); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func LeaveGroup(ctx context.Context, db *sql.DB, userID, groupID int) (LeaveGroupResult, error) {
	result := LeaveGroupResult{}

	tx, err := db.BeginTx(ctx, nil)
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

		if err := removeGroupChatParticipantTx(ctx, tx, groupID, userID); err != nil {
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

	if err := removeGroupChatParticipantTx(ctx, tx, groupID, userID); err != nil {
		return result, err
	}

	if err := tx.Commit(); err != nil {
		return result, err
	}

	result.OwnerTransferred = true
	result.NewOwnerID = newOwnerID
	return result, nil
}

func PromoteGroupModerator(ctx context.Context, db *sql.DB, ownerID, groupID, targetUserID int) error {
	tx, err := db.BeginTx(ctx, nil)
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

func DemoteGroupModerator(ctx context.Context, db *sql.DB, ownerID, groupID, targetUserID int) error {
	tx, err := db.BeginTx(ctx, nil)
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

func UpdateGroupSettings(ctx context.Context, db *sql.DB, ownerID, groupID int, visibility, joinMode *string) (GroupSettingsResult, error) {
	out := GroupSettingsResult{GroupID: groupID}

	tx, err := db.BeginTx(ctx, nil)
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

func GetGroupSettings(ctx context.Context, db *sql.DB, ownerID, groupID int) (GroupSettingsResult, error) {
	out := GroupSettingsResult{GroupID: groupID}

	var actualOwnerID int
	if err := db.QueryRowContext(ctx, `
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

func GetActiveGroupMembers(ctx context.Context, db *sql.DB, groupID int) ([]models.GroupMemberListItem, error) {
	var groupExists int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM groups WHERE id = ?`, groupID).Scan(&groupExists); err != nil {
		return nil, err
	}
	if groupExists == 0 {
		return nil, ErrGroupNotFound
	}

	rows, err := db.QueryContext(ctx, `
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

func isGroupModeratorOrOwner(ctx context.Context, db *sql.DB, groupID, userID int) (bool, error) {
	var role string
	err := db.QueryRowContext(ctx, `
		SELECT role
		FROM group_members
		WHERE group_id = ? AND user_id = ? AND status = 'active'
	`, groupID, userID).Scan(&role)
	if err == nil {
		return role == "owner" || role == "moderator", nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}

	var groupExists int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM groups WHERE id = ?`, groupID).Scan(&groupExists); err != nil {
		return false, err
	}
	if groupExists == 0 {
		return false, ErrGroupNotFound
	}

	return false, nil
}

func GetPendingGroupJoinRequests(ctx context.Context, db *sql.DB, actorID, groupID int) ([]models.GroupPendingItem, error) {
	allowed, err := isGroupModeratorOrOwner(ctx, db, groupID, actorID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrNotGroupModeratorOrOwner
	}

	rows, err := db.QueryContext(ctx, `
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

func GetPendingGroupInvites(ctx context.Context, db *sql.DB, actorID, groupID int) ([]models.GroupPendingItem, error) {
	allowed, err := isGroupModeratorOrOwner(ctx, db, groupID, actorID)
	if err != nil {
		return nil, err
	}
	if !allowed {
		return nil, ErrNotGroupModeratorOrOwner
	}

	rows, err := db.QueryContext(ctx, `
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

func GetGroupPageView(ctx context.Context, db *sql.DB, viewerID, groupID int) (GroupPageView, error) {
	var out GroupPageView

	if err := db.QueryRowContext(ctx, `
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
	err := db.QueryRowContext(ctx, `
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
	err = db.QueryRowContext(ctx, `
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

func RemovePendingGroupInvite(ctx context.Context, db *sql.DB, actorID, groupID, targetUserID int) error {
	tx, err := db.BeginTx(ctx, nil)
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

func RemoveOwnPendingGroupRequest(ctx context.Context, db *sql.DB, userID, groupID int) error {
	tx, err := db.BeginTx(ctx, nil)
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
