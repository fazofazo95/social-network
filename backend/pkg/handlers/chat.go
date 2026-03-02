package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/repository"
	"backend/pkg/responses"
	"backend/pkg/services"
)

type ChatHandler struct {
	Service services.ChatService
}

func NewChatHandler(s services.ChatService) *ChatHandler {
    return &ChatHandler{Service: s}
}

type markChatReadInput struct {
	LastMessageID int `json:"last_message_id"`
}

func (h *ChatHandler) RegisterRoutes(mux *http.ServeMux) {
	auth := middleware.WithAuth
	
	mux.Handle("GET /api/chats", middleware.Chain(h.ListChatsHandler, auth))
	mux.Handle("GET /api/chats/{chat_id}/messages", middleware.Chain(h.GetChatMessagesHandler, auth))
	mux.Handle("POST /api/chats/direct/{user_id}/messages", middleware.Chain(h.SendDirectMessageHandler, auth))
	mux.Handle("POST /api/chats/{chat_id}/read", middleware.Chain(h.MarkChatReadHandler, auth))
	mux.Handle("POST /api/groups/{id}/chat/messages", middleware.Chain(h.SendGroupMessageHandler, auth))
}

func (h *ChatHandler)  SendDirectMessageHandler(w http.ResponseWriter, r *http.Request) {
	senderID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	targetID, err := strconv.Atoi(r.PathValue("user_id"))
	if err != nil || targetID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid target user id")
		return
	}

	var in models.SendMessageInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	message, err := h.Service.SendDirectMessage(r.Context(), senderID, targetID, in)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInvalidChatMessage):
			responses.SendError(w, http.StatusBadRequest, "invalid message payload")
			return
		case errors.Is(err, repository.ErrDirectChatNotAllowed):
			responses.SendError(w, http.StatusForbidden, "direct chat is not allowed for these users")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to send direct message: "+err.Error())
			return
		}
	}
	responses.SendCreated(w, "message sent", message)
}


func (h *ChatHandler) SendGroupMessageHandler(w http.ResponseWriter, r *http.Request) {
	senderID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	groupID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || groupID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid group id")
		return
	}

	var in models.SendMessageInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		responses.SendError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	message, err := h.Service.SendGroupMessage(r.Context(), senderID, groupID, in)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInvalidChatMessage):
			responses.SendError(w, http.StatusBadRequest, "invalid message payload")
			return
		case errors.Is(err, repository.ErrGroupChatNotFound), errors.Is(err, repository.ErrGroupNotFound):
			responses.SendError(w, http.StatusNotFound, "group chat not found")
			return
		case errors.Is(err, repository.ErrChatForbidden):
			responses.SendError(w, http.StatusForbidden, "only active group members can send group chat messages")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to send group message: "+err.Error())
			return
		}
	}

	responses.SendCreated(w, "message sent", message)
}

func (h *ChatHandler) ListChatsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	limit := 30
	if s := r.URL.Query().Get("limit"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v <= 0 {
			responses.SendError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = v
	}

	offset := 0
	if s := r.URL.Query().Get("offset"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v < 0 {
			responses.SendError(w, http.StatusBadRequest, "invalid offset")
			return
		}
		offset = v
	}

	items, err := h.Service.ListChats(r.Context(), userID, limit, offset)
	if err != nil {
		responses.SendError(w, http.StatusInternalServerError, "failed to list chats: "+err.Error())
		return
	}

	responses.SendSuccess(w, "chats fetched successfully", map[string]interface{}{
		"items":  items,
		"limit":  limit,
		"offset": offset,
	})
}

func (h *ChatHandler) GetChatMessagesHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	chatID, err := strconv.Atoi(r.PathValue("chat_id"))
	if err != nil || chatID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid chat id")
		return
	}

	limit := 30
	if s := r.URL.Query().Get("limit"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v <= 0 {
			responses.SendError(w, http.StatusBadRequest, "invalid limit")
			return
		}
		limit = v
	}

	beforeID := 0
	if s := r.URL.Query().Get("before_id"); s != "" {
		v, err := strconv.Atoi(s)
		if err != nil || v <= 0 {
			responses.SendError(w, http.StatusBadRequest, "invalid before_id")
			return
		}
		beforeID = v
	}

	items, err := h.Service.GetChatMessages(r.Context(), userID, chatID, beforeID, limit)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrChatNotFound):
			responses.SendError(w, http.StatusNotFound, "chat not found")
			return
		case errors.Is(err, repository.ErrChatForbidden):
			responses.SendError(w, http.StatusForbidden, "chat access forbidden")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to fetch chat messages: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "messages fetched successfully", map[string]interface{}{
		"chat_id":   chatID,
		"items":     items,
		"limit":     limit,
		"before_id": beforeID,
	})
}

func (h *ChatHandler) MarkChatReadHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		responses.SendError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	chatID, err := strconv.Atoi(r.PathValue("chat_id"))
	if err != nil || chatID <= 0 {
		responses.SendError(w, http.StatusBadRequest, "invalid chat id")
		return
	}

	var in markChatReadInput
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&in)
	}

	if err := h.Service.MarkChatRead(r.Context(), userID, chatID, in.LastMessageID); err != nil {
		switch {
		case errors.Is(err, repository.ErrChatNotFound):
			responses.SendError(w, http.StatusNotFound, "chat not found")
			return
		case errors.Is(err, repository.ErrChatForbidden):
			responses.SendError(w, http.StatusForbidden, "chat access forbidden")
			return
		case errors.Is(err, repository.ErrInvalidChatMessage):
			responses.SendError(w, http.StatusBadRequest, "invalid last_message_id")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to mark chat as read: "+err.Error())
			return
		}
	}

	responses.SendSuccess(w, "chat marked as read", map[string]interface{}{
		"chat_id":         chatID,
		"last_message_id": in.LastMessageID,
	})
}
