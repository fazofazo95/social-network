package ws

import (
	"backend/pkg/models"
	"sync"

	"github.com/gorilla/websocket"
)

type BroadcastMessage struct {
	Payload      any
	RecipientIDs []int
}

type Client struct {
	Hub    *Hub
	UserID int
	Conn   *websocket.Conn
	Send   chan any
}

type Hub struct {
	Clients    map[int]*Client
	Broadcast  chan BroadcastMessage
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[int]*Client),
		Broadcast:  make(chan BroadcastMessage),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.handleRegister(client)

		case client := <-h.Unregister:
			h.handleUnregister(client)

		case bMsg := <-h.Broadcast:
			h.routeMessage(bMsg)
		}
	}
}

func (h *Hub) handleRegister(client *Client) {
	h.mu.Lock()
	h.Clients[client.UserID] = client
	
	onlineUserIDs := make([]int, 0, len(h.Clients))
	for id := range h.Clients {
		onlineUserIDs = append(onlineUserIDs, id)
	}
	h.mu.Unlock()

	client.Send <- map[string]any{
		"event":           "presence_snapshot",
		"online_user_ids": onlineUserIDs,
	}

	h.broadcastPresence(client.UserID, true)
}

func (h *Hub) handleUnregister(client *Client) {
	h.mu.Lock()
	if _, ok := h.Clients[client.UserID]; ok {
		delete(h.Clients, client.UserID)
		close(client.Send)
	}
	h.mu.Unlock()

	h.broadcastPresence(client.UserID, false)
}

func (h *Hub) routeMessage(bMsg BroadcastMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, userID := range bMsg.RecipientIDs {
		if client, ok := h.Clients[userID]; ok {
			select {
			case client.Send <- bMsg.Payload:
			default:
				go func(c *Client) { h.Unregister <- c }(client)
			}
		}
	}
}

func (h *Hub) broadcastPresence(userID int, online bool) {
	payload := map[string]any{
		"event":   "presence",
		"user_id": userID,
		"online":  online,
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.Clients {
		if client.UserID == userID {
			continue
		}
		
		select {
		case client.Send <- payload:
		default:
		}
	}
}

func (h *Hub) SendChatMessage(msg *models.ChatMessage, participantIDs []int) {
	h.Broadcast <- BroadcastMessage{
		Payload:      msg,
		RecipientIDs: participantIDs,
	}
}