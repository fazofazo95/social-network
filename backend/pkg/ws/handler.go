package ws

import (
	"backend/pkg/middleware"
	"backend/pkg/models"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,

	CheckOrigin: func(r *http.Request) bool {
		return true 
	},
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		log.Println("Unauthorized WS connection attempt")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	client := &Client{
		Hub:    hub,
		UserID: userID,
		Conn:   conn,
		Send:   make(chan models.ChatMessage, 256),
	}

	hub.Register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			if err := c.Conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024)

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
