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
	SearchGroups(ctx context.Context, userID int, query string, limit int) ([]models.SearchGroupItem, error)
	GetActiveGroupsForUser(ctx context.Context, userID int) ([]models.GroupActiveItem, error)
	GetUserPendingGroupRequests(ctx context.Context, userID int) ([]models.GroupUserPendingItem, error)
	GetUserPendingGroupInvites(ctx context.Context, userID int) ([]models.GroupUserPendingItem, error)

	// Events
	CreateGroupEvent(ctx context.Context, actorID, groupID int, in models.GroupEventCreateInput) (*models.GroupEvent, error)
	GetGroupEventsTimeline(ctx context.Context, actorID, groupID int) (models.GroupEventsTimeline, error)
	RespondToGroupEventInvite(ctx context.Context, actorID, groupID, eventID int, reactionType string) error
	ChangeGroupEventResponse(ctx context.Context, actorID, groupID, eventID int, reactionType string) error
	DeleteGroupEvent(ctx context.Context, actorID, groupID, eventID int) error

	// Group Posts
	CreateGroupPost(ctx context.Context, actorID, groupID int, content, image string) (*models.Post, error)
	GetGroupPosts(ctx context.Context, viewerID, groupID, page int) ([]*models.Post, error)
	DeleteGroupPost(ctx context.Context, actorID, groupID, postID int) error
}

type groupService struct {
	repo     repository.GroupRepository
	notifSvc NotificationService
}

func NewGroupService(r repository.GroupRepository, ns NotificationService) GroupService {
	return &groupService{
		repo:     r,
		notifSvc: ns,
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

	status, err := s.repo.RequestToJoinGroup(ctx, userID, groupID)
	if err != nil {
		return "", err
	}

	if status == "requested" && s.notifSvc != nil {
		if err := s.notifSvc.NotifyGroupJoinRequest(ctx, userID, groupID); err != nil {
			return "", err
		}
	}

	return status, nil
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

	status, err := s.repo.InviteUserToGroup(ctx, inviterID, groupID, targetUserID)
	if err != nil {
		return "", err
	}

	if status == "requested" && s.notifSvc != nil {
		if err := s.notifSvc.NotifyGroupInvite(ctx, inviterID, groupID, targetUserID); err != nil {
			return "", err
		}
	}

	return status, nil
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

func (s *groupService) SearchGroups(ctx context.Context, userID int, query string, limit int) ([]models.SearchGroupItem, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []models.SearchGroupItem{}, nil
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 25 {
		limit = 25
	}
	return s.repo.SearchGroups(ctx, userID, query, limit)
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

	created, err := s.repo.CreateGroupEvent(ctx, actorID, groupID, in)
	if err != nil {
		return nil, err
	}

	if s.notifSvc != nil {
		if err := s.notifSvc.NotifyGroupEventCreated(ctx, actorID, groupID, created.ID, created.Title); err != nil {
			return nil, err
		}
	}

	return created, nil
}

func (s *groupService) GetGroupEventsTimeline(ctx context.Context, actorID, groupID int) (models.GroupEventsTimeline, error) {
	if groupID <= 0 {
		return models.GroupEventsTimeline{}, errors.New("invalid group id")
	}
	return s.repo.GetGroupEventsTimeline(ctx, actorID, groupID)
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

// --- Group Posts ---
func (s *groupService) CreateGroupPost(ctx context.Context, actorID, groupID int, content, image string) (*models.Post, error) {
	content = strings.TrimSpace(content)
	if content == "" && image == "" {
		return nil, errors.New("post content or image is required")
	}
	if groupID <= 0 {
		return nil, errors.New("invalid group id")
	}
	return s.repo.CreateGroupPost(ctx, actorID, groupID, content, image)
}

func (s *groupService) GetGroupPosts(ctx context.Context, viewerID, groupID, page int) ([]*models.Post, error) {
	if groupID <= 0 {
		return nil, errors.New("invalid group id")
	}
	if page < 1 {
		page = 1
	}
	limit := 10
	offset := (page - 1) * limit
	return s.repo.GetGroupPosts(ctx, viewerID, groupID, limit, offset)
}

func (s *groupService) DeleteGroupPost(ctx context.Context, actorID, groupID, postID int) error {
	if groupID <= 0 || postID <= 0 {
		return errors.New("invalid group or post id")
	}
	return s.repo.DeleteGroupPost(ctx, actorID, groupID, postID)
}
