package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/middleware"
	"backend/pkg/models"
	"backend/pkg/responses"
	websocket "backend/pkg/ws"
)

type markChatReadInput struct {
	LastMessageID int `json:"last_message_id"`
}

func SendDirectMessageHandler(hub *websocket.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		message, err := queries.SendDirectMessage(r.Context(), database.DB, senderID, targetID, in)
		if err != nil {
			switch {
			case errors.Is(err, queries.ErrInvalidChatMessage):
				responses.SendError(w, http.StatusBadRequest, "invalid message payload")
				return
			case errors.Is(err, queries.ErrDirectChatNotAllowed):
				responses.SendError(w, http.StatusForbidden, "direct chat is not allowed for these users")
				return
			default:
				responses.SendError(w, http.StatusInternalServerError, "failed to send direct message: "+err.Error())
				return
			}
		}
		hub.Broadcast <- message
		responses.SendCreated(w, "message sent", message)
	}
}

func SendGroupMessageHandler(w http.ResponseWriter, r *http.Request) {
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

	message, err := queries.SendGroupMessage(r.Context(), database.DB, senderID, groupID, in)
	if err != nil {
		switch {
		case errors.Is(err, queries.ErrInvalidChatMessage):
			responses.SendError(w, http.StatusBadRequest, "invalid message payload")
			return
		case errors.Is(err, queries.ErrGroupChatNotFound), errors.Is(err, queries.ErrGroupNotFound):
			responses.SendError(w, http.StatusNotFound, "group chat not found")
			return
		case errors.Is(err, queries.ErrChatForbidden):
			responses.SendError(w, http.StatusForbidden, "only active group members can send group chat messages")
			return
		default:
			responses.SendError(w, http.StatusInternalServerError, "failed to send group message: "+err.Error())
			return
		}
	}

	responses.SendCreated(w, "message sent", message)
}

func ListChatsHandler(w http.ResponseWriter, r *http.Request) {
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

	items, err := queries.ListChats(r.Context(), database.DB, userID, limit, offset)
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

func GetChatMessagesHandler(w http.ResponseWriter, r *http.Request) {
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

	items, err := queries.GetChatMessages(r.Context(), database.DB, userID, chatID, beforeID, limit)
	if err != nil {
		switch {
		case errors.Is(err, queries.ErrChatNotFound):
			responses.SendError(w, http.StatusNotFound, "chat not found")
			return
		case errors.Is(err, queries.ErrChatForbidden):
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

func MarkChatReadHandler(w http.ResponseWriter, r *http.Request) {
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

	if err := queries.MarkChatRead(r.Context(), database.DB, userID, chatID, in.LastMessageID); err != nil {
		switch {
		case errors.Is(err, queries.ErrChatNotFound):
			responses.SendError(w, http.StatusNotFound, "chat not found")
			return
		case errors.Is(err, queries.ErrChatForbidden):
			responses.SendError(w, http.StatusForbidden, "chat access forbidden")
			return
		case errors.Is(err, queries.ErrInvalidChatMessage):
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
