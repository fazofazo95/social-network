package services

import (
	"backend/pkg/repository"
	"context"
)

type ReactionService interface {
	AddReaction(ctx context.Context, userID, postID int) (int, error)
	RemoveReaction(ctx context.Context, userID, postID int) (int, error)
}

type reactionService struct {
	repo repository.ReactionRepository
}

func NewReactionService(r repository.ReactionRepository) ReactionService {
	return &reactionService{
		repo: r,
	}
}

func (s *reactionService) AddReaction(ctx context.Context, userID, postID int) (int, error) {
	likeCount, err := s.repo.AddReaction(ctx, userID, postID)
	if err != nil {
		return 0, err
	}

	return likeCount, nil
}

func (s *reactionService) RemoveReaction(ctx context.Context, userID, postID int) (int, error) {
	likeCount, err := s.repo.RemoveReaction(ctx, userID, postID)
	if err != nil {
		return 0, err
	}

	return likeCount, nil
}
