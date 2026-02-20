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
	// Επιτρέπουμε το origin από το frontend μας
	CheckOrigin: func(r *http.Request) bool {
		return true // Στο production βάλε το domain σου
	},
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// 1. Authentication: Παίρνουμε το UserID από το context (που έβαλε το middleware σου)
	userID, err := middleware.UserIDFromContext(r.Context())
	if err != nil {
		log.Println("Unauthorized WS connection attempt")
		return
	}

	// 2. Upgrade: Μετατροπή της HTTP σε WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Upgrade error: %v", err)
		return
	}

	// 3. Δημιουργία Client
	client := &Client{
		Hub:    hub,
		UserID: userID,
		Conn:   conn,
		Send:   make(chan models.ChatMessage, 256),
	}

	// 4. Εγγραφή στο Hub
	hub.Register <- client

	// 5. Έναρξη των Goroutines (Οι "πνεύμονες" της σύνδεσης)
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
				// Το Hub έκλεισε το κανάλι
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// Στέλνουμε το ChatMessage ως JSON στον browser
			if err := c.Conn.WriteJSON(message); err != nil {
				return
			}

		case <-ticker.C:
			// Heartbeat: Στέλνουμε Ping για να μη κλείσει το TCP από timeout
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

	// Ρυθμίσεις ορίων για ασφάλεια
	c.Conn.SetReadLimit(512 * 1024) // 512KB limit

	for {
		// Εδώ περιμένουμε μηνύματα από τον client
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			// Αν κλείσει το tab ή κοπεί το net, βγαίνουμε από το loop
			break
		}
		// Αν θέλεις ο χρήστης να στέλνει μηνύματα ΜΟΝΟ μέσω WebSocket
		// (και όχι μέσω POST), η λογική θα έμπαινε εδώ.
		// Αλλά εμείς έχουμε τους HTTP Handlers ως triggers!
	}
}
