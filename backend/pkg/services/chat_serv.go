package services

import (
	"backend/pkg/models"
	"backend/pkg/repository"
	"backend/pkg/ws"
	"context"
	"errors"
	"log"
	"strings"
)

type ChatService interface {
	SendDirectMessage(ctx context.Context, senderID, targetUserID int, in models.SendMessageInput) (*models.ChatMessage, error)
	SendGroupMessage(ctx context.Context, senderID, groupID int, in models.SendMessageInput) (*models.ChatMessage, error)

	GetChatMessages(ctx context.Context, userID, chatID, beforeID, limit int) ([]*models.ChatMessage, error)
	MarkChatRead(ctx context.Context, userID, chatID, lastMessageID int) error

	ListChats(ctx context.Context, userID, limit, offset int) ([]*models.ChatSummary, error)
}

type chatService struct {
	repo repository.ChatRepository
	hub  *ws.Hub
}

func NewChatService(r repository.ChatRepository, h *ws.Hub) ChatService {
	return &chatService{
		repo: r,
		hub:  h,
	}
}

func (s *chatService) SendDirectMessage(ctx context.Context, senderID, targetUserID int, in models.SendMessageInput) (*models.ChatMessage, error) {
	if senderID == targetUserID {
		return nil, repository.ErrDirectChatNotAllowed
	}

	validatedIn, err := s.validateMessageInput(in)
	if err != nil {
		log.Printf("invalid message input: %v", err)
		return nil, err
	}

	allowed, err := s.repo.CanUsersChat(ctx, senderID, targetUserID)
	if err != nil {
		log.Printf("error checking chat permission: %v", err)
		return nil, err
	}
	if !allowed {
		return nil, repository.ErrDirectChatNotAllowed
	}

	lowID, highID := s.normalizePair(senderID, targetUserID)

	chatID, err := s.repo.GetDirectChatID(ctx, lowID, highID)
	if err != nil {
		log.Printf("error fetching direct chat ID: %v", err)
		if errors.Is(err, repository.ErrChatNotFound) {
			chatID, err = s.repo.CreateDirectChat(ctx, lowID, highID, senderID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	message, err := s.repo.SaveMessage(ctx, chatID, senderID, validatedIn)
	if err != nil {
		log.Printf("error saving message: %v", err)
		return nil, err
	}

	s.hub.Broadcast <- ws.BroadcastMessage{
		Payload:      message,
		RecipientIDs: []int{senderID, targetUserID},
	}

	return message, nil
}

func (s *chatService) SendGroupMessage(ctx context.Context, senderID, groupID int, in models.SendMessageInput) (*models.ChatMessage, error) {
	validatedIn, err := s.validateMessageInput(in)
	if err != nil {
		return nil, err
	}

	chatID, err := s.repo.GetChatIDByGroupID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	isActive, err := s.repo.IsGroupMemberActive(ctx, groupID, senderID)
	if err != nil {
		return nil, err
	}
	if !isActive {
		return nil, repository.ErrChatForbidden
	}

	if err := s.repo.EnsureParticipant(ctx, chatID, senderID); err != nil {
		return nil, err
	}

	message, err := s.repo.SaveMessage(ctx, chatID, senderID, validatedIn)
	if err != nil {
		return nil, err
	}

	participants, err := s.repo.GetGroupParticipantIDs(ctx, groupID)
	if err != nil {
		log.Printf("[ERROR] Failed to fetch participants for broadcast: %v", err)
		return message, nil
	}

	s.hub.Broadcast <- ws.BroadcastMessage{
		Payload:      message,
		RecipientIDs: participants,
	}

	return message, nil
}

func (s *chatService) GetChatMessages(ctx context.Context, userID, chatID, beforeID, limit int) ([]*models.ChatMessage, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}

	canAccess, err := s.repo.UserHasChatAccess(ctx, userID, chatID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, repository.ErrChatForbidden
	}

	messages, err := s.repo.FetchMessages(ctx, chatID, beforeID, limit)
	if err != nil {
		return nil, err
	}

	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

func (s *chatService) MarkChatRead(ctx context.Context, userID, chatID, lastMessageID int) error {
	canAccess, err := s.repo.UserHasChatAccess(ctx, userID, chatID)
	if err != nil {
		return err
	}
	if !canAccess {
		return repository.ErrChatForbidden
	}

	if lastMessageID <= 0 {
		lastMessageID, err = s.repo.GetLatestMessageID(ctx, chatID)
		if err != nil {
			return err
		}
	} else {
		msgChatID, err := s.repo.GetMessageChatID(ctx, lastMessageID)
		if err != nil {
			return err
		}
		if msgChatID != chatID {
			return repository.ErrInvalidChatMessage
		}
	}

	return s.repo.UpdateLastReadMessage(ctx, userID, chatID, lastMessageID)
}

func (s *chatService) ListChats(ctx context.Context, userID, limit, offset int) ([]*models.ChatSummary, error) {
	// 1. Sanitization (Business Rules)
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	if offset < 0 {
		offset = 0
	}

	// 2. Data Retrieval
	return s.repo.FetchChatSummaries(ctx, userID, limit, offset)
}

func (s *chatService) normalizePair(a, b int) (int, int) {
	if a < b {
		return a, b
	}
	return b, a
}

func (s *chatService) validateMessageInput(in models.SendMessageInput) (models.SendMessageInput, error) {
	in.MessageType = strings.ToLower(strings.TrimSpace(in.MessageType))
	in.Body = strings.TrimSpace(in.Body)
	in.MediaURL = strings.TrimSpace(in.MediaURL)

	if in.MessageType == "" {
		in.MessageType = "text"
	}

	switch in.MessageType {
	case "text":
		if in.Body == "" {
			return models.SendMessageInput{}, repository.ErrInvalidChatMessage
		}
	case "image":
		if in.MediaURL == "" {
			return models.SendMessageInput{}, repository.ErrInvalidChatMessage
		}
	case "text_image":
		if in.Body == "" || in.MediaURL == "" {
			return models.SendMessageInput{}, repository.ErrInvalidChatMessage
		}
	default:
		return models.SendMessageInput{}, repository.ErrInvalidChatMessage
	}

	return in, nil
}
