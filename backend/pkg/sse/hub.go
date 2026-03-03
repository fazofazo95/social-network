package sse

import (
	"backend/pkg/models"
	"sync"
)

type Hub struct {
	mu      sync.RWMutex
	clients map[int]map[chan models.NotificationSSEPayload]struct{}
}

func NewHub() *Hub {
	return &Hub{
		clients: make(map[int]map[chan models.NotificationSSEPayload]struct{}),
	}
}

func (h *Hub) Subscribe(userID int) chan models.NotificationSSEPayload {
	ch := make(chan models.NotificationSSEPayload, 32)

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[userID]; !ok {
		h.clients[userID] = make(map[chan models.NotificationSSEPayload]struct{})
	}

	h.clients[userID][ch] = struct{}{}
	return ch
}

func (h *Hub) Unsubscribe(userID int, ch chan models.NotificationSSEPayload) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if subscribers, ok := h.clients[userID]; ok {
		if _, exists := subscribers[ch]; exists {
			delete(subscribers, ch)
			close(ch)
		}
		if len(subscribers) == 0 {
			delete(h.clients, userID)
		}
	}
}

func (h *Hub) Publish(userID int, payload models.NotificationSSEPayload) {
	h.mu.RLock()
	subscribers, ok := h.clients[userID]
	h.mu.RUnlock()
	if !ok {
		return
	}

	for ch := range subscribers {
		select {
		case ch <- payload:
		default:
		}
	}
}
