package services

import (
	"backend/pkg/models"
	"backend/pkg/repository"
	"context"
	"errors"
)

type FollowService interface {
	FollowUser(ctx context.Context, req models.FollowRequest) (string, error)
	DeleteFollow(ctx context.Context, followerID, targetID int) (int64, error)
	RemoveFollower(ctx context.Context, currentUserID, targetFollowerID int) (int64, error)
	AcceptFollow(ctx context.Context, followerID, followedID int) (int64, error)
	RejectFollow(ctx context.Context, followerID, followedID int) (int64, error)
	BlockFollow(ctx context.Context, blockerID, targetID int) (int64, error)
	UnblockFollow(ctx context.Context, blockerID, targetID int) (int64, error)

	GetFollowingUsers(ctx context.Context, target, userID int) ([]*models.FollowListUser, error)
	GetFollowers(ctx context.Context, userID int) ([]*models.FollowListUser, error)
	GetFollowingByUser(ctx context.Context, viewerID, targetID int) ([]*models.FollowListUser, error)
	GetFollowersByUser(ctx context.Context, viewerID, targetID int) ([]*models.FollowListUser, error)
	GetBlockedUsers(ctx context.Context, userID int) ([]*models.FollowListUser, error)
	GetPendingRequests(ctx context.Context, userID int) ([]*models.FollowListUser, error)
}

type followService struct {
	repo        repository.FollowRepository
	profileRepo repository.ProfileRepository
}

func NewFollowService(r repository.FollowRepository, pr repository.ProfileRepository) FollowService {
	return &followService{
		repo:        r,
		profileRepo: pr,
	}
}

func (s *followService) FollowUser(ctx context.Context, req models.FollowRequest) (string, error) {
	if req.FollowedID == 0 || req.FollowerID == 0 {
		return "", errors.New("follower_id and followed_id are required")
	}

	isUserPrivate, err := s.profileRepo.UserPrivacy(ctx, req.FollowedID)
	if err != nil {
		return "", err
	}

	status := "pending"
	if !isUserPrivate {
		status = "accepted"
	}

	err = s.repo.CreateFollow(ctx, req, status)
	if err != nil {
		return "", err
	}

	return "nil", nil
}

func (s *followService) DeleteFollow(ctx context.Context, followerID, targetID int) (int64, error) {
	if followerID == 0 || targetID == 0 {
		return 0, errors.New("follower_id and followed_id are required")
	}
	return s.repo.DeleteFollow(ctx, followerID, targetID)
}

func (s *followService) RemoveFollower(ctx context.Context, currentUserID, targetFollowerID int) (int64, error) {
	if currentUserID == 0 || targetFollowerID == 0 {
		return 0, errors.New("current_user_id and target_follower_id are required")
	}
	return s.repo.RemoveFollower(ctx, currentUserID, targetFollowerID)
}

func (s *followService) AcceptFollow(ctx context.Context, followerID, followedID int) (int64, error) {
	if followerID == 0 || followedID == 0 {
		return 0, errors.New("follower_id and followed_id are required")
	}
	return s.repo.AcceptFollow(ctx, followerID, followedID)
}

func (s *followService) RejectFollow(ctx context.Context, followerID, followedID int) (int64, error) {
	if followerID == 0 || followedID == 0 {
		return 0, errors.New("follower_id and followed_id are required")
	}
	return s.repo.RejectFollow(ctx, followerID, followedID)
}

func (s *followService) BlockFollow(ctx context.Context, blockerID, targetID int) (int64, error) {
	if blockerID == 0 || targetID == 0 {
		return 0, errors.New("blocker_id and target_id are required")
	}
	return s.repo.BlockFollow(ctx, blockerID, targetID)
}

func (s *followService) UnblockFollow(ctx context.Context, blockerID, targetID int) (int64, error) {
	if blockerID == 0 || targetID == 0 {
		return 0, errors.New("blocker_id and target_id are required")
	}
	return s.repo.UnblockFollow(ctx, blockerID, targetID)
}

func (s *followService) GetFollowingUsers(ctx context.Context, target, userID int) ([]*models.FollowListUser, error) {
	if userID <= 0 || target <= 0 {
		return nil, errors.New("invalid user id")
	}

	if target == userID {
		return s.repo.GetFollowingUsers(ctx, userID)
	}

	return s.repo.GetFollowingUsersForViewer(ctx, target, userID)
}

func (s *followService) GetFollowers(ctx context.Context, userID int) ([]*models.FollowListUser, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}

	users, err := s.repo.GetFollowers(ctx, userID)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *followService) GetFollowingByUser(ctx context.Context, viewerID, targetID int) ([]*models.FollowListUser, error) {
	if viewerID <= 0 || targetID <= 0 {
		return nil, errors.New("invalid user id")
	}

	if targetID == viewerID {
		return s.repo.GetFollowers(ctx, targetID)
	}

	return s.repo.GetFollowersForViewer(ctx, targetID, viewerID)
}

func (s *followService) GetFollowersByUser(ctx context.Context, viewerID, targetID int) ([]*models.FollowListUser, error) {
	if viewerID <= 0 || targetID <= 0 {
		return nil, errors.New("invalid user id")
	}

	if targetID == viewerID {
		return s.repo.GetFollowers(ctx, targetID)
	}
	return s.repo.GetFollowersForViewer(ctx, viewerID, targetID)
}

func (s *followService) GetBlockedUsers(ctx context.Context, userID int) ([]*models.FollowListUser, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}

	users, err := s.repo.GetBlockedUsers(ctx, userID)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (s *followService) GetPendingRequests(ctx context.Context, userID int) ([]*models.FollowListUser, error) {
	if userID <= 0 {
		return nil, errors.New("invalid user id")
	}

	users, err := s.repo.GetPendingIncomingRequests(ctx, userID)
	if err != nil {
		return nil, err
	}

	return users, nil
}
