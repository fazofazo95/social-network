package services

import (
	"backend/pkg/models"
	"backend/pkg/repository"
	"backend/pkg/sse"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type NotificationService interface {
	ListByUser(ctx context.Context, userID, limit, offset int) ([]models.NotificationWithActor, error)
	MarkSeen(ctx context.Context, userID, notificationID int) error
	MarkAllSeen(ctx context.Context, userID int) (int64, error)
	GetByIDForUser(ctx context.Context, userID, notificationID int) (models.NotificationWithActor, error)
	SetStatusForUser(ctx context.Context, userID, notificationID int, status string) error
	NotifyFollowRequest(ctx context.Context, followerID, followedID int) error
	NotifyGroupInvite(ctx context.Context, inviterID, groupID, targetUserID int) error
	NotifyGroupJoinRequest(ctx context.Context, requesterID, groupID int) error
	NotifyGroupEventCreated(ctx context.Context, creatorID, groupID, eventID int, title string) error
}

type notificationService struct {
	repo repository.NotificationRepository
	hub  *sse.Hub
}

func NewNotificationService(repo repository.NotificationRepository, hub *sse.Hub) NotificationService {
	return &notificationService{repo: repo, hub: hub}
}

func (s *notificationService) ListByUser(ctx context.Context, userID, limit, offset int) ([]models.NotificationWithActor, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.ListByRecipient(ctx, userID, limit, offset)
}

func (s *notificationService) MarkSeen(ctx context.Context, userID, notificationID int) error {
	if userID <= 0 || notificationID <= 0 {
		return errors.New("invalid user or notification id")
	}
	return s.repo.MarkSeen(ctx, notificationID, userID)
}

func (s *notificationService) MarkAllSeen(ctx context.Context, userID int) (int64, error) {
	if userID <= 0 {
		return 0, errors.New("invalid user id")
	}
	return s.repo.MarkAllSeen(ctx, userID)
}

func (s *notificationService) GetByIDForUser(ctx context.Context, userID, notificationID int) (models.NotificationWithActor, error) {
	if userID <= 0 || notificationID <= 0 {
		return models.NotificationWithActor{}, errors.New("invalid user or notification id")
	}

	return s.repo.GetByIDForRecipient(ctx, notificationID, userID)
}

func (s *notificationService) SetStatusForUser(ctx context.Context, userID, notificationID int, status string) error {
	if userID <= 0 || notificationID <= 0 {
		return errors.New("invalid user or notification id")
	}

	status = strings.ToLower(strings.TrimSpace(status))
	if status != models.NotificationStatusAccepted && status != models.NotificationStatusRejected {
		return errors.New("invalid notification status")
	}

	return s.repo.MarkSeenAndStatus(ctx, notificationID, userID, status)
}

func (s *notificationService) NotifyFollowRequest(ctx context.Context, followerID, followedID int) error {
	if followerID <= 0 || followedID <= 0 {
		return errors.New("invalid follower or followed id")
	}

	followerName, err := s.repo.GetUserDisplayName(ctx, followerID)
	if err != nil {
		return err
	}

	metadata, _ := json.Marshal(map[string]any{
		"followed_id": followedID,
		"follower_id": followerID,
	})

	n, err := s.repo.Create(ctx, models.CreateNotificationInput{
		RecipientID: followedID,
		ActorID:     intPtr(followerID),
		Type:        models.NotificationTypeFollowRequest,
		Content:     fmt.Sprintf("%s sent you a follow request", followerName),
		Metadata:    string(metadata),
	})
	if err != nil {
		return err
	}

	s.publish(n)
	return nil
}

func (s *notificationService) NotifyGroupInvite(ctx context.Context, inviterID, groupID, targetUserID int) error {
	if inviterID <= 0 || groupID <= 0 || targetUserID <= 0 {
		return errors.New("invalid inviter, group or target user id")
	}

	inviterName, err := s.repo.GetUserDisplayName(ctx, inviterID)
	if err != nil {
		return err
	}

	groupName, err := s.repo.GetGroupNameByID(ctx, groupID)
	if err != nil {
		return err
	}

	metadata, _ := json.Marshal(map[string]any{
		"group_id":   groupID,
		"inviter_id": inviterID,
	})

	n, err := s.repo.Create(ctx, models.CreateNotificationInput{
		RecipientID: targetUserID,
		ActorID:     intPtr(inviterID),
		Type:        models.NotificationTypeGroupInvite,
		GroupID:     intPtr(groupID),
		Content:     fmt.Sprintf("%s invited you to join group %s", inviterName, groupName),
		Metadata:    string(metadata),
	})
	if err != nil {
		return err
	}

	s.publish(n)
	return nil
}

func (s *notificationService) NotifyGroupJoinRequest(ctx context.Context, requesterID, groupID int) error {
	if requesterID <= 0 || groupID <= 0 {
		return errors.New("invalid requester or group id")
	}

	ownerID, err := s.repo.GetGroupOwnerID(ctx, groupID)
	if err != nil {
		return err
	}

	requesterName, err := s.repo.GetUserDisplayName(ctx, requesterID)
	if err != nil {
		return err
	}

	groupName, err := s.repo.GetGroupNameByID(ctx, groupID)
	if err != nil {
		return err
	}

	metadata, _ := json.Marshal(map[string]any{
		"group_id":       groupID,
		"requester_id":   requesterID,
		"group_owner_id": ownerID,
	})

	n, err := s.repo.Create(ctx, models.CreateNotificationInput{
		RecipientID: ownerID,
		ActorID:     intPtr(requesterID),
		Type:        models.NotificationTypeGroupJoin,
		GroupID:     intPtr(groupID),
		Content:     fmt.Sprintf("%s requested to join group %s", requesterName, groupName),
		Metadata:    string(metadata),
	})
	if err != nil {
		return err
	}

	s.publish(n)
	return nil
}

func (s *notificationService) NotifyGroupEventCreated(ctx context.Context, creatorID, groupID, eventID int, title string) error {
	if creatorID <= 0 || groupID <= 0 || eventID <= 0 {
		return errors.New("invalid creator, group or event id")
	}

	creatorName, err := s.repo.GetUserDisplayName(ctx, creatorID)
	if err != nil {
		return err
	}

	memberIDs, err := s.repo.GetGroupActiveMemberIDs(ctx, groupID, creatorID)
	if err != nil {
		return err
	}
	if len(memberIDs) == 0 {
		return nil
	}

	groupName, err := s.repo.GetGroupNameByID(ctx, groupID)
	if err != nil {
		return err
	}

	for _, memberID := range memberIDs {
		metadata, _ := json.Marshal(map[string]any{
			"group_id":   groupID,
			"event_id":   eventID,
			"creator_id": creatorID,
		})

		n, err := s.repo.Create(ctx, models.CreateNotificationInput{
			RecipientID: memberID,
			ActorID:     intPtr(creatorID),
			Type:        models.NotificationTypeGroupEvent,
			GroupID:     intPtr(groupID),
			EventID:     intPtr(eventID),
			Content:     fmt.Sprintf("%s created a new event in group %s: %s", creatorName, groupName, title),
			Metadata:    string(metadata),
		})
		if err != nil {
			return err
		}

		s.publish(n)
	}

	return nil
}

func (s *notificationService) publish(n *models.NotificationWithActor) {
	if s.hub == nil || n == nil {
		return
	}

	s.hub.Publish(n.RecipientID, models.NotificationSSEPayload{
		Event:        "notification:new",
		Notification: *n,
	})
}

func intPtr(value int) *int {
	v := value
	return &v
}
