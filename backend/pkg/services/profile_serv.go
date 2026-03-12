package services

import (
	"backend/pkg/models"
	"backend/pkg/repository"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type ProfileService interface {
	GetUserProfileView(ctx context.Context, viewerID int, targetID string) (*models.UserProfileDTO, error)
	GetSocialStatus(ctx context.Context, viewerID, targetID int) (string, error)
	UpdateUserMedia(ctx context.Context, targetID int, imageURL, coverImageURL string) error
	DiscoveredUser(ctx context.Context, userID, limit int) ([]*models.DiscoveredUser, error)
	SearchUsers(ctx context.Context, userID int, query string, limit int) ([]models.SearchUserItem, error)
	GetUserVisibilitySettings(ctx context.Context, userID int) (*models.VisibilitySettings, error)
	UpdateVisibility(ctx context.Context, userID int, req models.UpdateVisibilityRequest) (*models.VisibilitySettings, error)
	GetUserContentSettings(ctx context.Context, userID int) (*models.UserProfileDTO, error)
	UpdateUserContent(ctx context.Context, userID int, req models.UserProfileRequest) (*models.UserProfileDTO, error)
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
		var err error
		targetIDInt, err = strconv.Atoi(targetID)
		if err != nil || targetIDInt <= 0 {
			log.Printf("invalid targetID: %v", targetID)
			return nil, fmt.Errorf("invalid user id")
		}
	}

	rawProfile, err := s.repo.GetRawUserProfile(ctx, viewerID, targetIDInt)
	if err != nil {
		log.Printf("error fetching raw profile: %v", err)
		return nil, err
	}

	socialStatus, err := s.GetSocialStatus(ctx, viewerID, targetIDInt)
	if err != nil {
		log.Printf("error fetching social status: %v", err)
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

	isPrivate := raw.ProfileType == true
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

func (s *profileService) SearchUsers(ctx context.Context, userID int, query string, limit int) ([]models.SearchUserItem, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []models.SearchUserItem{}, nil
	}
	if limit <= 0 {
		limit = 10
	}
	if limit > 25 {
		limit = 25
	}

	users, err := s.repo.SearchUsers(ctx, userID, query, limit)
	if err != nil {
		return nil, err
	}

	// Safety guard: never return the requester in search results.
	filtered := make([]models.SearchUserItem, 0, len(users))
	for _, u := range users {
		if u.ID == userID {
			continue
		}
		filtered = append(filtered, u)
	}

	return filtered, nil
}

func (s *profileService) GetUserVisibilitySettings(ctx context.Context, userID int) (*models.VisibilitySettings, error) {
	raw, err := s.repo.GetVisibilityRaw(ctx, userID)
	if err != nil {
		return nil, err
	}

	mapVis := func(v int) string {
		if v == 1 {
			return "visible"
		}
		return "hidden"
	}

	settings := &models.VisibilitySettings{
		EmailVis:              mapVis(raw.EmailVis),
		BirthdayDateVis:       mapVis(raw.BirthdayVis),
		RelationshipStatusVis: mapVis(raw.RelVis),
		EmployedAtVis:         mapVis(raw.EmployedVis),
		PhoneNumberVis:        mapVis(raw.PhoneVis),
		AboutMeVis:            mapVis(raw.AboutVis),
		NicknameVis:           mapVis(raw.NickVis),
		FollowVis:             mapVis(raw.FollowVis),
	}

	if raw.ProfileType == true {
		settings.ProfileType = "private"
	} else {
		settings.ProfileType = "public"
	}

	return settings, nil
}

func (s *profileService) UpdateVisibility(ctx context.Context, userID int, req models.UpdateVisibilityRequest) (*models.VisibilitySettings, error) {

	parse := func(s *string) *int {
		if s == nil {
			return nil
		}
		val := 0
		str := strings.ToLower(*s)
		if str == "visible" || str == "true" || str == "public" || str == "private" {
			if str == "visible" || str == "true" || str == "private" {
				val = 1
			}
			return &val
		}
		return &val
	}

	err := s.repo.UpdateUserVisibilitySettings(ctx, userID,
		parse(req.EmailVis),
		parse(req.BirthdayVis),
		parse(req.RelationshipStatusVis),
		parse(req.EmployedAtVis),
		parse(req.PhoneNumberVis),
		parse(req.AboutMeVis),
		parse(req.NicknameVis),
		parse(req.FollowVis),
		parse(req.ProfileType),
	)
	if err != nil {
		return nil, err
	}

	return s.GetUserVisibilitySettings(ctx, userID)
}

func (s *profileService) GetUserContentSettings(ctx context.Context, userID int) (*models.UserProfileDTO, error) {
	raw, err := s.repo.GetContentRaw(ctx, userID)
	if err != nil {
		return nil, err
	}

	dto := &models.UserProfileDTO{
		ID:                 raw.ID,
		FirstName:          raw.FirstName,
		LastName:           raw.LastName,
		BirthdayDate:       raw.BirthdayDate,
		RelationshipStatus: raw.RelationshipStatus,
		EmployedAt:         raw.EmployedAt,
		Location:           raw.Location,
		PhoneNumber:        raw.PhoneNumber,
		Nickname:           raw.Nickname,
		AboutMe:            raw.AboutMe,
	}

	return dto, nil
}

func (s *profileService) UpdateUserContent(ctx context.Context, userID int, req models.UserProfileRequest) (*models.UserProfileDTO, error) {
	if req.Level == "" {
		req.Level = "user"
	}

	var birthdayStr *string
	if req.Birthday != nil {
		s := req.Birthday.Format("2006-01-02")
		birthdayStr = &s
	}

	err := s.repo.UpdateProfileContent(ctx, userID, req, birthdayStr)
	if err != nil {
		return nil, err
	}

	return s.GetUserContentSettings(ctx, userID)
}
