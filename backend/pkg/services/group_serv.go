package services

import (
	"context"
	"errors"
	"strings"

	"backend/pkg/models"
	"backend/pkg/repository"
)

type GroupService interface {
	// Group CRUD
	CreateGroup(ctx context.Context, ownerID int, in models.CreateGroupInput) (*models.GroupResponse, error)
	DeleteGroup(ctx context.Context, requesterID, groupID int) error

	// Join/Leave operations
	RequestToJoinGroup(ctx context.Context, userID, groupID int) (string, error)
	AcceptGroupJoinRequest(ctx context.Context, approverID, groupID, requesterID int) error
	RejectGroupJoinRequest(ctx context.Context, approverID, groupID, requesterID int) error
	InviteUserToGroup(ctx context.Context, inviterID, groupID, targetUserID int) (string, error)
	AcceptGroupInvite(ctx context.Context, userID, groupID int) error
	RejectGroupInvite(ctx context.Context, userID, groupID int) error
	KickGroupMember(ctx context.Context, actorID, groupID, targetUserID int) error
	LeaveGroup(ctx context.Context, userID, groupID int) (repository.LeaveGroupResult, error)

	// Member management
	PromoteGroupModerator(ctx context.Context, ownerID, groupID, targetUserID int) error
	DemoteGroupModerator(ctx context.Context, ownerID, groupID, targetUserID int) error
	GetActiveGroupMembers(ctx context.Context, groupID int) ([]models.GroupMemberListItem, error)

	// Settings
	UpdateGroupSettings(ctx context.Context, ownerID, groupID int, visibility, joinMode *string) (repository.GroupSettingsResult, error)
	GetGroupSettings(ctx context.Context, ownerID, groupID int) (repository.GroupSettingsResult, error)

	// Pending requests/invites
	GetPendingGroupJoinRequests(ctx context.Context, actorID, groupID int) ([]models.GroupPendingItem, error)
	GetPendingGroupInvites(ctx context.Context, actorID, groupID int) ([]models.GroupPendingItem, error)
	RemovePendingGroupInvite(ctx context.Context, actorID, groupID, targetUserID int) error
	RemoveOwnPendingGroupRequest(ctx context.Context, userID, groupID int) error

	// Group listing
	GetGroupPageView(ctx context.Context, viewerID, groupID int) (repository.GroupPageView, error)
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

type groupService struct {
	repo repository.GroupRepository
}

func NewGroupService(r repository.GroupRepository) GroupService {
	return &groupService{
		repo: r,
	}
}

// --- Group CRUD ---
func (s *groupService) CreateGroup(ctx context.Context, ownerID int, in models.CreateGroupInput) (*models.GroupResponse, error) {
	if in.Name == "" {
		return nil, errors.New("name is required")
	}
	if in.Visibility != "public" && in.Visibility != "private" {
		return nil, errors.New("visibility must be public or private")
	}
	switch in.JoinMode {
	case "auto", "request", "invite", "request_and_invite":
	default:
		return nil, errors.New("join_mode must be auto, request, invite, or request_and_invite")
	}

	return s.repo.CreateGroup(ctx, ownerID, in)
}

func (s *groupService) DeleteGroup(ctx context.Context, requesterID, groupID int) error {
	if groupID <= 0 {
		return errors.New("invalid group id")
	}

	return s.repo.DeleteGroup(ctx, requesterID, groupID)
}

// --- Join/Leave Operations ---
func (s *groupService) RequestToJoinGroup(ctx context.Context, userID, groupID int) (string, error) {
	if groupID <= 0 {
		return "", errors.New("invalid group id")
	}
	return s.repo.RequestToJoinGroup(ctx, userID, groupID)
}

func (s *groupService) AcceptGroupJoinRequest(ctx context.Context, approverID, groupID, requesterID int) error {
	if groupID <= 0 || requesterID <= 0 {
		return errors.New("invalid group or requester id")
	}
	return s.repo.AcceptGroupJoinRequest(ctx, approverID, groupID, requesterID)
}

func (s *groupService) RejectGroupJoinRequest(ctx context.Context, approverID, groupID, requesterID int) error {
	if groupID <= 0 || requesterID <= 0 {
		return errors.New("invalid group or requester id")
	}
	return s.repo.RejectGroupJoinRequest(ctx, approverID, groupID, requesterID)
}

func (s *groupService) InviteUserToGroup(ctx context.Context, inviterID, groupID, targetUserID int) (string, error) {
	if groupID <= 0 || targetUserID <= 0 {
		return "", errors.New("invalid group or target user id")
	}
	return s.repo.InviteUserToGroup(ctx, inviterID, groupID, targetUserID)
}

func (s *groupService) AcceptGroupInvite(ctx context.Context, userID, groupID int) error {
	if groupID <= 0 {
		return errors.New("invalid group id")
	}
	return s.repo.AcceptGroupInvite(ctx, userID, groupID)
}

func (s *groupService) RejectGroupInvite(ctx context.Context, userID, groupID int) error {
	if groupID <= 0 {
		return errors.New("invalid group id")
	}
	return s.repo.RejectGroupInvite(ctx, userID, groupID)
}

func (s *groupService) KickGroupMember(ctx context.Context, actorID, groupID, targetUserID int) error {
	if groupID <= 0 || targetUserID <= 0 {
		return errors.New("invalid group or target user id")
	}
	return s.repo.KickGroupMember(ctx, actorID, groupID, targetUserID)
}

func (s *groupService) LeaveGroup(ctx context.Context, userID, groupID int) (repository.LeaveGroupResult, error) {
	if groupID <= 0 {
		return repository.LeaveGroupResult{}, errors.New("invalid group id")
	}
	return s.repo.LeaveGroup(ctx, userID, groupID)
}

// --- Member Management ---
func (s *groupService) PromoteGroupModerator(ctx context.Context, ownerID, groupID, targetUserID int) error {
	if groupID <= 0 || targetUserID <= 0 {
		return errors.New("invalid group or target user id")
	}
	return s.repo.PromoteGroupModerator(ctx, ownerID, groupID, targetUserID)
}

func (s *groupService) DemoteGroupModerator(ctx context.Context, ownerID, groupID, targetUserID int) error {
	if groupID <= 0 || targetUserID <= 0 {
		return errors.New("invalid group or target user id")
	}
	return s.repo.DemoteGroupModerator(ctx, ownerID, groupID, targetUserID)
}

func (s *groupService) GetActiveGroupMembers(ctx context.Context, groupID int) ([]models.GroupMemberListItem, error) {
	if groupID <= 0 {
		return nil, errors.New("invalid group id")
	}
	return s.repo.GetActiveGroupMembers(ctx, groupID)
}

// --- Settings ---
func (s *groupService) UpdateGroupSettings(ctx context.Context, ownerID, groupID int, visibility, joinMode *string) (repository.GroupSettingsResult, error) {
	if groupID <= 0 {
		return repository.GroupSettingsResult{}, errors.New("invalid group id")
	}
	return s.repo.UpdateGroupSettings(ctx, ownerID, groupID, visibility, joinMode)
}

func (s *groupService) GetGroupSettings(ctx context.Context, ownerID, groupID int) (repository.GroupSettingsResult, error) {
	if groupID <= 0 {
		return repository.GroupSettingsResult{}, errors.New("invalid group id")
	}
	return s.repo.GetGroupSettings(ctx, ownerID, groupID)
}

// --- Pending Requests/Invites ---
func (s *groupService) GetPendingGroupJoinRequests(ctx context.Context, actorID, groupID int) ([]models.GroupPendingItem, error) {
	if groupID <= 0 {
		return nil, errors.New("invalid group id")
	}
	return s.repo.GetPendingGroupJoinRequests(ctx, actorID, groupID)
}

func (s *groupService) GetPendingGroupInvites(ctx context.Context, actorID, groupID int) ([]models.GroupPendingItem, error) {
	if groupID <= 0 {
		return nil, errors.New("invalid group id")
	}
	return s.repo.GetPendingGroupInvites(ctx, actorID, groupID)
}

func (s *groupService) RemovePendingGroupInvite(ctx context.Context, actorID, groupID, targetUserID int) error {
	if groupID <= 0 || targetUserID <= 0 {
		return errors.New("invalid group or target user id")
	}
	return s.repo.RemovePendingGroupInvite(ctx, actorID, groupID, targetUserID)
}

func (s *groupService) RemoveOwnPendingGroupRequest(ctx context.Context, userID, groupID int) error {
	if groupID <= 0 {
		return errors.New("invalid group id")
	}
	return s.repo.RemoveOwnPendingGroupRequest(ctx, userID, groupID)
}

// --- Group Listing ---
func (s *groupService) GetGroupPageView(ctx context.Context, viewerID, groupID int) (repository.GroupPageView, error) {
	if groupID <= 0 {
		return repository.GroupPageView{}, errors.New("invalid group id")
	}
	return s.repo.GetGroupPageView(ctx, viewerID, groupID)
}

func (s *groupService) DiscoverGroups(ctx context.Context, userID, limit, offset int) ([]models.GroupDiscoverItem, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.DiscoverGroups(ctx, userID, limit, offset)
}

func (s *groupService) GetActiveGroupsForUser(ctx context.Context, userID int) ([]models.GroupActiveItem, error) {
	return s.repo.GetActiveGroupsForUser(ctx, userID)
}

func (s *groupService) GetUserPendingGroupRequests(ctx context.Context, userID int) ([]models.GroupUserPendingItem, error) {
	return s.repo.GetUserPendingGroupRequests(ctx, userID)
}

func (s *groupService) GetUserPendingGroupInvites(ctx context.Context, userID int) ([]models.GroupUserPendingItem, error) {
	return s.repo.GetUserPendingGroupInvites(ctx, userID)
}

// --- Events ---
func (s *groupService) CreateGroupEvent(ctx context.Context, actorID, groupID int, in models.GroupEventCreateInput) (*models.GroupEvent, error) {
	in.Title = strings.TrimSpace(in.Title)
	in.Description = strings.TrimSpace(in.Description)
	in.EventDay = strings.TrimSpace(in.EventDay)
	in.EventTime = strings.TrimSpace(in.EventTime)

	if in.Title == "" {
		return nil, errors.New("title is required")
	}
	if in.EventDay == "" {
		return nil, errors.New("event_day is required")
	}
	if in.EventTime == "" {
		return nil, errors.New("event_time is required")
	}

	if groupID <= 0 {
		return nil, errors.New("invalid group id")
	}

	return s.repo.CreateGroupEvent(ctx, actorID, groupID, in)
}

func (s *groupService) GetGroupEventInviteableMembers(ctx context.Context, actorID, groupID, eventID int) ([]models.GroupMemberListItem, error) {
	if groupID <= 0 || eventID <= 0 {
		return nil, errors.New("invalid group or event id")
	}
	return s.repo.GetGroupEventInviteableMembers(ctx, actorID, groupID, eventID)
}

func (s *groupService) InviteGroupEventMember(ctx context.Context, actorID, groupID, eventID, targetUserID int) error {
	if groupID <= 0 || eventID <= 0 || targetUserID <= 0 {
		return errors.New("invalid group, event or target user id")
	}
	return s.repo.InviteGroupEventMember(ctx, actorID, groupID, eventID, targetUserID)
}

func (s *groupService) InviteAllGroupEventMembers(ctx context.Context, actorID, groupID, eventID int) (int, error) {
	if groupID <= 0 || eventID <= 0 {
		return 0, errors.New("invalid group or event id")
	}
	return s.repo.InviteAllGroupEventMembers(ctx, actorID, groupID, eventID)
}

func (s *groupService) RespondToGroupEventInvite(ctx context.Context, actorID, groupID, eventID int, reactionType string) error {
	reactionType = strings.TrimSpace(strings.ToLower(reactionType))
	if reactionType != "going" && reactionType != "not_going" {
		return errors.New("reaction_type must be going or not_going")
	}
	if groupID <= 0 || eventID <= 0 {
		return errors.New("invalid group or event id")
	}
	return s.repo.RespondToGroupEventInvite(ctx, actorID, groupID, eventID, reactionType)
}

func (s *groupService) ChangeGroupEventResponse(ctx context.Context, actorID, groupID, eventID int, reactionType string) error {
	reactionType = strings.TrimSpace(strings.ToLower(reactionType))
	if reactionType != "going" && reactionType != "not_going" {
		return errors.New("reaction_type must be going or not_going")
	}
	if groupID <= 0 || eventID <= 0 {
		return errors.New("invalid group or event id")
	}
	return s.repo.ChangeGroupEventResponse(ctx, actorID, groupID, eventID, reactionType)
}

func (s *groupService) DeleteGroupEvent(ctx context.Context, actorID, groupID, eventID int) error {
	if groupID <= 0 || eventID <= 0 {
		return errors.New("invalid group or event id")
	}
	return s.repo.DeleteGroupEvent(ctx, actorID, groupID, eventID)
}
