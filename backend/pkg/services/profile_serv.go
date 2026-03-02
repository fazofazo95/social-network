package services

import (
	"backend/pkg/models"
	"backend/pkg/repository"
	"context"
	"fmt"
	"strconv"
)

type ProfileService interface {
	GetUserProfileView(ctx context.Context, viewerID int, targetID string) (*models.UserProfileDTO, error)
	GetSocialStatus(ctx context.Context, viewerID, targetID int) (string, error)
	UpdateUserMedia(ctx context.Context, targetID int, imageURL, coverImageURL string) error
	DiscoveredUser(ctx context.Context, userID, limit int) ([]*models.DiscoveredUser, error)
}

type profileService struct {
	repo repository.ProfileRepository
}

func NewProfileService(r repository.ProfileRepository) ProfileService {
	return &profileService{
		repo: r,
	}
}

func (s *profileService) GetUserProfileView(ctx context.Context, viewerID int, targetID string) (*models.UserProfileDTO, error) {
	var targetIDInt int
	if targetID == "me" {
		targetIDInt = viewerID
	} else {
		targetIDInt, err := strconv.Atoi(targetID)
		if err != nil || targetIDInt <= 0 {
			return nil, fmt.Errorf("invalid user id")
		}
	}

	rawProfile, err := s.repo.GetRawUserProfile(ctx, viewerID, targetIDInt)
	if err != nil {
		return nil, err
	}

	socialStatus, err := s.GetSocialStatus(ctx, viewerID, targetIDInt)
	if err != nil {
		return nil, err
	}

	profileDTO := mapRawToDTO(rawProfile, viewerID, targetIDInt, socialStatus)
	return profileDTO, nil
}

func (s *profileService) GetSocialStatus(ctx context.Context, viewerID, targetID int) (string, error) {
	outgoing, err := s.repo.GetRelation(ctx, viewerID, targetID)
	if err != nil {
		return "", err
	}

	incoming, err := s.repo.GetRelation(ctx, targetID, viewerID)
	if err != nil {
		return "", err
	}

	if outgoing == "blocked" {
		return "Blocked", nil
	}
	if incoming == "blocked" {
		return "You_Are_Blocked", nil
	}

	switch outgoing {
	case "accepted":
		return "Following", nil
	case "pending":
		return "Pending", nil
	}

	if incoming == "accepted" {
		return "Follow Back", nil
	}

	return "Follow", nil
}

func (s *profileService) UpdateUserMedia(ctx context.Context, targetID int, imageURL, coverImageURL string) error {
	if imageURL == "" && coverImageURL == "" {
		return fmt.Errorf("no avatar or cover file provided")
	}

	err := s.repo.UpdateUserMedia(ctx, targetID, imageURL, coverImageURL)
	if err != nil {
		return err
	}
	return nil
}

func mapRawToDTO(raw *models.RawProfileData, viewerID, targetID int, currentStatus string) *models.UserProfileDTO {
	dto := &models.UserProfileDTO{
		ID:             raw.ID,
		FirstName:      raw.FirstName,
		LastName:       raw.LastName,
		ProfilePicture: raw.ProfilePicture,
		CoverImage:     raw.CoverImage,
		Followers:      raw.FollowersCount,
		Following:      raw.FollowingCount,
		OwnProfile:     viewerID == targetID,
		CurrentStatus:  currentStatus,
	}

	mapVis := func(v int) string {
		if v == 1 {
			return "visible"
		}
		return "hidden"
	}

	if dto.OwnProfile {
		dto.FollowVis = "visible"
		dto.Location = raw.Location
		dto.Email = raw.Email
		dto.BirthdayDate = raw.BirthdayDate
		dto.RelationshipStatus = raw.RelationshipStatus
		dto.EmployedAt = raw.EmployedAt
		dto.PhoneNumber = raw.PhoneNumber
		dto.Nickname = raw.Nickname
		dto.AboutMe = raw.AboutMe
		return dto
	}

	if currentStatus == "Blocked" || currentStatus == "You_Are_Blocked" {
		return &models.UserProfileDTO{
			ID:            raw.ID,
			OwnProfile:    false,
			CurrentStatus: currentStatus,
		}
	}

	isPrivate := raw.ProfileType == 1
	canViewFull := !isPrivate || currentStatus == "Following"

	if !canViewFull {
		dto.FollowVis = "hidden"
		return dto
	}

	dto.FollowVis = mapVis(raw.FollowVis)

	if raw.EmailVis == 1 {
		dto.Email = raw.Email
	}
	if raw.BirthdayVis == 1 {
		dto.BirthdayDate = raw.BirthdayDate
	}
	if raw.RelationshipVis == 1 {
		dto.RelationshipStatus = raw.RelationshipStatus
	}
	if raw.EmployedVis == 1 {
		dto.EmployedAt = raw.EmployedAt
		dto.Location = raw.Location
	}
	if raw.PhoneVis == 1 {
		dto.PhoneNumber = raw.PhoneNumber
	}
	if raw.NicknameVis == 1 {
		dto.Nickname = raw.Nickname
	}
	if raw.AboutVis == 1 {
		dto.AboutMe = raw.AboutMe
	}

	return dto
}

func (s *profileService) DiscoveredUser(ctx context.Context, userID, limit int) ([]*models.DiscoveredUser, error) {
	return s.repo.DiscoverUsers(ctx, userID, limit)
}
