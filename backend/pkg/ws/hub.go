package ws

import (
	"backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/models"
	"context"
	"database/sql"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Hub struct {
	Clients    map[int]*Client
	Broadcast  chan models.ChatMessage
	Register   chan *Client
	Unregister chan *Client
	DB         *sql.DB
	mu         sync.Mutex
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[int]*Client),
		Broadcast:  make(chan models.ChatMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		DB:         database.DB,
	}
}

type Client struct {
	Hub    *Hub
	UserID int
	Conn   *websocket.Conn
	Send   chan any
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.UserID] = client
			h.sendPresenceSnapshotLocked(client)
			h.broadcastPresenceLocked(client.UserID, true, client.UserID)
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.UserID]; ok {
				delete(h.Clients, client.UserID)
				close(client.Send)
				h.broadcastPresenceLocked(client.UserID, false, client.UserID)
			}
			h.mu.Unlock()

		case msg := <-h.Broadcast:
			h.routeMessage(msg)
		}
	}
}

func (h *Hub) sendPresenceSnapshotLocked(client *Client) {
	onlineUserIDs := make([]int, 0, len(h.Clients))
	for userID := range h.Clients {
		onlineUserIDs = append(onlineUserIDs, userID)
	}

	h.sendToClientLocked(client, map[string]any{
		"event":           "presence_snapshot",
		"online_user_ids": onlineUserIDs,
	})
}

func (h *Hub) broadcastPresenceLocked(userID int, online bool, excludeUserID int) {
	payload := map[string]any{
		"event":   "presence",
		"user_id": userID,
		"online":  online,
	}

	for id, client := range h.Clients {
		if excludeUserID > 0 && id == excludeUserID {
			continue
		}
		h.sendToClientLocked(client, payload)
	}
}

func (h *Hub) sendToClientLocked(client *Client, payload any) {
	select {
	case client.Send <- payload:
	default:
		close(client.Send)
		delete(h.Clients, client.UserID)
	}
}

func (h *Hub) routeMessage(msg models.ChatMessage) {
	h.mu.Lock()
	defer h.mu.Unlock()

	participants := h.getParticipants(msg.ChatID)

	for _, userID := range participants {
		if userID == msg.SenderID {
			continue
		}

		if client, ok := h.Clients[userID]; ok {
			select {
			case client.Send <- msg:
			default:
				close(client.Send)
				delete(h.Clients, userID)
			}
		}
	}
}

func (h *Hub) getParticipants(chatID int) []int {
	participants, err := queries.GetChatParticipants(context.Background(), h.DB, chatID)
	if err != nil {
		log.Printf("Error getting participants for chat %d: %v", chatID, err)
		return nil
	}
	return participants
}
