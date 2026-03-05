package repository

import (
	"backend/pkg/models"
	"context"
	"database/sql"
	"errors"
	"strings"
)

var (
	ErrGroupNameTaken               = errors.New("group name already in use")
	ErrGroupNotFound                = errors.New("group not found")
	ErrGroupEventNotFound           = errors.New("group event not found")
	ErrNotGroupOwner                = errors.New("only the group owner can delete the group")
	ErrPrivateGroup                 = errors.New("cannot join private group from this endpoint")
	ErrInviteOnlyGroup              = errors.New("group is invite-only")
	ErrAlreadyGroupMember           = errors.New("user is already a group member")
	ErrAlreadyRequestedToJoin       = errors.New("user has already requested to join")
	ErrAlreadyInvitedToGroup        = errors.New("user has already been invited")
	ErrNotGroupModeratorOrOwner     = errors.New("only group owner or moderators can approve requests")
	ErrGroupJoinRequestNotFound     = errors.New("group join request not found")
	ErrGroupInviteNotFound          = errors.New("group invite not found")
	ErrGroupMemberNotFound          = errors.New("group member not found")
	ErrCannotKickGroupStaff         = errors.New("cannot kick owner or moderator")
	ErrGroupMemberRoleMismatch      = errors.New("group member role mismatch")
	ErrGroupMemberIsActive          = errors.New("group member is active")
	ErrNotActiveGroupMember         = errors.New("user is not an active group member")
	ErrTargetNotActiveMember        = errors.New("target user is not an active group member")
	ErrCannotInviteSelf             = errors.New("cannot invite yourself")
	ErrGroupEventAlreadyAnswered    = errors.New("user already invited or responded to event")
	ErrNotInvitedToEvent            = errors.New("user is not invited or responded to event")
	ErrGroupEventAlreadyResponded   = errors.New("user already responded to event")
	ErrGroupEventNoResponseToChange = errors.New("no event response to change")
	ErrGroupEventResponseUnchanged  = errors.New("event response already set")
)

type GroupRepository interface {
	// Group CRUD
	CreateGroup(ctx context.Context, ownerID int, in models.CreateGroupInput) (*models.GroupResponse, error)
	DeleteGroup(ctx context.Context, requesterID, groupID int) error
	GetGroupOwner(ctx context.Context, groupID int) (int, error)

	// Join/Leave operations
	RequestToJoinGroup(ctx context.Context, userID, groupID int) (string, error)
	AcceptGroupJoinRequest(ctx context.Context, approverID, groupID, requesterID int) error
	RejectGroupJoinRequest(ctx context.Context, approverID, groupID, requesterID int) error
	InviteUserToGroup(ctx context.Context, inviterID, groupID, targetUserID int) (string, error)
	AcceptGroupInvite(ctx context.Context, userID, groupID int) error
	RejectGroupInvite(ctx context.Context, userID, groupID int) error
	KickGroupMember(ctx context.Context, actorID, groupID, targetUserID int) error
	LeaveGroup(ctx context.Context, userID, groupID int) (LeaveGroupResult, error)

	// Member management
	PromoteGroupModerator(ctx context.Context, ownerID, groupID, targetUserID int) error
	DemoteGroupModerator(ctx context.Context, ownerID, groupID, targetUserID int) error
	GetActiveGroupMembers(ctx context.Context, groupID int) ([]models.GroupMemberListItem, error)

	// Settings
	UpdateGroupSettings(ctx context.Context, ownerID, groupID int, visibility, joinMode *string) (GroupSettingsResult, error)
	GetGroupSettings(ctx context.Context, ownerID, groupID int) (GroupSettingsResult, error)

	// Pending requests/invites
	GetPendingGroupJoinRequests(ctx context.Context, actorID, groupID int) ([]models.GroupPendingItem, error)
	GetPendingGroupInvites(ctx context.Context, actorID, groupID int) ([]models.GroupPendingItem, error)
	RemovePendingGroupInvite(ctx context.Context, actorID, groupID, targetUserID int) error
	RemoveOwnPendingGroupRequest(ctx context.Context, userID, groupID int) error

	// Group listing
	GetGroupPageView(ctx context.Context, viewerID, groupID int) (GroupPageView, error)
	DiscoverGroups(ctx context.Context, userID, limit, offset int) ([]models.GroupDiscoverItem, error)
	GetActiveGroupsForUser(ctx context.Context, userID int) ([]models.GroupActiveItem, error)
	GetUserPendingGroupRequests(ctx context.Context, userID int) ([]models.GroupUserPendingItem, error)
	GetUserPendingGroupInvites(ctx context.Context, userID int) ([]models.GroupUserPendingItem, error)

	// Events
	CreateGroupEvent(ctx context.Context, actorID, groupID int, in models.GroupEventCreateInput) (*models.GroupEvent, error)
	GetGroupEventInviteableMembers(ctx context.Context, actorID, groupID, eventID int) ([]models.GroupMemberListItem, error)
	InviteGroupEventMember(ctx context.Context, actorID, groupID, eventID, targetUserID int) error
	InviteAllGroupEventMembers(ctx context.Context, actorID, groupID, eventID int) (int, error)
	RespondToGroupEventInvite(ctx context.Context, actorID, groupID, eventID int, reactionType string) error
	ChangeGroupEventResponse(ctx context.Context, actorID, groupID, eventID int, reactionType string) error
	DeleteGroupEvent(ctx context.Context, actorID, groupID, eventID int) error
}

// Result types
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

type sqliteGroupRepo struct {
	db *sql.DB
}

func NewGroupRepository(db *sql.DB) GroupRepository {
	return &sqliteGroupRepo{db: db}
}

func (r *sqliteGroupRepo) ensureGroupChatIDTx(ctx context.Context, tx *sql.Tx, groupID int) (int, error) {
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

func (r *sqliteGroupRepo) ensureGroupChatParticipantTx(ctx context.Context, tx *sql.Tx, groupID, userID int) error {
	chatID, err := r.ensureGroupChatIDTx(ctx, tx, groupID)
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

func (r *sqliteGroupRepo) removeGroupChatParticipantTx(ctx context.Context, tx *sql.Tx, groupID, userID int) error {
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

func (r *sqliteGroupRepo) CreateGroup(ctx context.Context, ownerID int, in models.CreateGroupInput) (*models.GroupResponse, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO groups (name, description, owner_id, visibility, group_picture,join_mode, group_members)
		VALUES (?, ?, ?, ?, ?, ?, 1)
	`, in.Name, in.Description, ownerID, in.Visibility, in.Picture, in.JoinMode)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "groups.name") {
			return nil, ErrGroupNameTaken
		}
		return nil, err
	}

	groupID64, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	groupID := int(groupID64)

	_, err = tx.ExecContext(ctx, `
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES (?, ?, 'owner', 'active')
	`, groupID, ownerID)
	if err != nil {
		return nil, err
	}

	chatRes, err := tx.ExecContext(ctx, `
		INSERT INTO chats (type, group_id, created_by, last_message_at)
		VALUES ('group', ?, ?, CURRENT_TIMESTAMP)
	`, groupID, ownerID)
	if err != nil {
		return nil, err
	}

	chatID64, err := chatRes.LastInsertId()
	if err != nil {
		return nil, err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO chat_participants (chat_id, user_id)
		VALUES (?, ?)
	`, int(chatID64), ownerID)
	if err != nil {
		return nil, err
	}

	out := &models.GroupResponse{
		ID:           groupID,
		Name:         in.Name,
		Description:  in.Description,
		OwnerID:      ownerID,
		Visibility:   in.Visibility,
		JoinMode:     in.JoinMode,
		GroupMembers: 1,
	}

	if err := tx.QueryRowContext(ctx, `SELECT COALESCE(datetime(created_at), '') FROM groups WHERE id = ?`, groupID).Scan(&out.CreatedAt); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return out, nil
}

func (r *sqliteGroupRepo) DeleteGroup(ctx context.Context, requesterID, groupID int) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

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

func (r *sqliteGroupRepo) GetGroupOwner(ctx context.Context, groupID int) (int, error) {
	var ownerID int
	err := r.db.QueryRowContext(ctx, `
		SELECT owner_id
		FROM groups
		WHERE id = ?
	`, groupID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, ErrGroupNotFound
		}
		return 0, err
	}

	return ownerID, nil
}

func (r *sqliteGroupRepo) RequestToJoinGroup(ctx context.Context, userID, groupID int) (string, error) {
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

		if err := r.ensureGroupChatParticipantTx(ctx, tx, groupID, userID); err != nil {
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
