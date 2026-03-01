package services

import (
	"backend/pkg/repository"
	"context"
)

type ProfileService interface {
	GetUserProfileView(ctx context.Context, viewerID int, targetID string) (map[string]interface{}, error)
}

type profileService struct {
	repo repository.ProfileRepository
}

func NewProfileService(r repository.ProfileRepository) ProfileService {
	return &profileService{
		repo: r,
	}
}

func (s *profileService) GetUserProfileView(ctx context.Context, viewerID int, targetID string) (map[string]interface{}, error){
	if targetID == "me"{
		
	}
}
