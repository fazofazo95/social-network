package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	database "backend/pkg/db/sqlite"
	handlers "backend/pkg/handlers"
)

type apiEnvelope struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type createGroupData struct {
	ID int `json:"id"`
}

type chatMessageData struct {
	ID       int `json:"id"`
	ChatID   int `json:"chat_id"`
	SenderID int `json:"sender_id"`
}

type chatsListData struct {
	Items []chatSummary `json:"items"`
}

type chatSummary struct {
	ChatID      int  `json:"chat_id"`
	Seen        bool `json:"seen"`
	OtherUserID int  `json:"other_user_id"`
	GroupID     int  `json:"group_id"`
}

func mustDo(client *http.Client, req *http.Request) (*http.Response, []byte) {
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("request failed: %v", err)
	}
	out, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	return resp, out
}

func login(client *http.Client, baseURL, email, password string) *http.Cookie {
	payload := map[string]string{"email": email, "password": password}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, out := mustDo(client, req)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("login failed for %s: %d %s", email, resp.StatusCode, string(out))
	}
	for _, c := range resp.Cookies() {
		if c.Name == "session_id" {
			return c
		}
	}
	log.Fatalf("session cookie missing for %s", email)
	return nil
}

func createGroup(client *http.Client, baseURL string, ownerCookie *http.Cookie, joinMode string) int {
	payload := map[string]string{
		"name":        fmt.Sprintf("Chat Integration %d", time.Now().UnixNano()),
		"description": "chat integration group",
		"visibility":  "public",
		"join_mode":   joinMode,
	}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/groups", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(ownerCookie)
	resp, out := mustDo(client, req)
	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("create group failed: %d %s", resp.StatusCode, string(out))
	}

	var env apiEnvelope
	if err := json.Unmarshal(out, &env); err != nil {
		log.Fatalf("decode create group envelope: %v", err)
	}
	var data createGroupData
	if err := json.Unmarshal(env.Data, &data); err != nil {
		log.Fatalf("decode create group data: %v", err)
	}
	return data.ID
}

func ensureAcceptedFollow(followerID, followedID int) {
	if _, err := database.DB.Exec(`
		INSERT INTO followers (follower_id, followed_id, status)
		VALUES (?, ?, 'accepted')
		ON CONFLICT(follower_id, followed_id)
		DO UPDATE SET status = 'accepted', updated_at = CURRENT_TIMESTAMP
	`, followerID, followedID); err != nil {
		log.Fatalf("ensure accepted follow %d->%d: %v", followerID, followedID, err)
	}
}

func sendDirectMessage(client *http.Client, baseURL string, senderCookie *http.Cookie, targetUserID int, body string) chatMessageData {
	payload := map[string]string{"message_type": "text", "body": body}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/chats/direct/%d/messages", baseURL, targetUserID), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(senderCookie)
	resp, out := mustDo(client, req)
	if resp.StatusCode != http.StatusCreated {
		log.Fatalf("send direct message failed: %d %s", resp.StatusCode, string(out))
	}

	var env apiEnvelope
	if err := json.Unmarshal(out, &env); err != nil {
		log.Fatalf("decode direct message envelope: %v", err)
	}
	var data chatMessageData
	if err := json.Unmarshal(env.Data, &data); err != nil {
		log.Fatalf("decode direct message data: %v", err)
	}
	return data
}

func listChats(client *http.Client, baseURL string, cookie *http.Cookie) chatsListData {
	req, _ := http.NewRequest(http.MethodGet, baseURL+"/api/chats?limit=50&offset=0", nil)
	req.AddCookie(cookie)
	resp, out := mustDo(client, req)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("list chats failed: %d %s", resp.StatusCode, string(out))
	}
	var env apiEnvelope
	if err := json.Unmarshal(out, &env); err != nil {
		log.Fatalf("decode list chats envelope: %v", err)
	}
	var data chatsListData
	if err := json.Unmarshal(env.Data, &data); err != nil {
		log.Fatalf("decode list chats data: %v", err)
	}
	return data
}

func markRead(client *http.Client, baseURL string, cookie *http.Cookie, chatID, lastMessageID int) {
	payload := map[string]int{"last_message_id": lastMessageID}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/chats/%d/read", baseURL, chatID), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	resp, out := mustDo(client, req)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("mark read failed: %d %s", resp.StatusCode, string(out))
	}
}

func findDirectChatSeen(items []chatSummary, otherUserID int) (bool, bool) {
	for _, item := range items {
		if item.OtherUserID == otherUserID {
			return item.Seen, true
		}
	}
	return false, false
}

func joinGroup(client *http.Client, baseURL string, cookie *http.Cookie, groupID int) {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/join", baseURL, groupID), nil)
	req.AddCookie(cookie)
	resp, out := mustDo(client, req)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("group join failed: %d %s", resp.StatusCode, string(out))
	}
}

func sendGroupMessage(client *http.Client, baseURL string, cookie *http.Cookie, groupID int, body string, expectedStatus int) chatMessageData {
	payload := map[string]string{"message_type": "text", "body": body}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/chat/messages", baseURL, groupID), bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	resp, out := mustDo(client, req)
	if resp.StatusCode != expectedStatus {
		log.Fatalf("send group message unexpected status: want=%d got=%d body=%s", expectedStatus, resp.StatusCode, string(out))
	}
	if expectedStatus != http.StatusCreated {
		return chatMessageData{}
	}

	var env apiEnvelope
	if err := json.Unmarshal(out, &env); err != nil {
		log.Fatalf("decode group message envelope: %v", err)
	}
	var data chatMessageData
	if err := json.Unmarshal(env.Data, &data); err != nil {
		log.Fatalf("decode group message data: %v", err)
	}
	return data
}

func leaveGroup(client *http.Client, baseURL string, cookie *http.Cookie, groupID int) {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/leave", baseURL, groupID), nil)
	req.AddCookie(cookie)
	resp, out := mustDo(client, req)
	if resp.StatusCode != http.StatusOK {
		log.Fatalf("leave group failed: %d %s", resp.StatusCode, string(out))
	}
}

func getMessagesExpect(client *http.Client, baseURL string, cookie *http.Cookie, chatID int, expectedStatus int) {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/chats/%d/messages?limit=20", baseURL, chatID), nil)
	req.AddCookie(cookie)
	resp, out := mustDo(client, req)
	if resp.StatusCode != expectedStatus {
		log.Fatalf("get chat messages unexpected status: want=%d got=%d body=%s", expectedStatus, resp.StatusCode, string(out))
	}
}

func main() {
	if err := database.Init("pkg/db/social_network.db"); err != nil {
		log.Fatalf("init db: %v", err)
	}
	defer database.DB.Close()

	mux := http.NewServeMux()
	handlers.UserRoutes(mux)
	handlers.AuthRoutes(mux)
	handlers.FollowRoutes(mux)
	handlers.GroupRoutes(mux)
	handlers.ChatRoutes(mux)
	handlers.PostRoutes(mux, database.DB)
	handlers.CommentRoutes(mux, database.DB)
	handlers.FeedRoutes(mux)

	ts := httptest.NewServer(mux)
	defer ts.Close()
	client := ts.Client()

	aliceCookie := login(client, ts.URL, "alice@example.com", "Password123!")
	bobCookie := login(client, ts.URL, "bob@example.com", "Password123!")
	carolCookie := login(client, ts.URL, "carol@example.com", "Password123!")

	// ---- Direct chat flow: send -> unseen for receiver -> mark read -> seen ----
	ensureAcceptedFollow(1, 2)
	directMsg := sendDirectMessage(client, ts.URL, aliceCookie, 2, "hello bob from alice")

	bobsChats := listChats(client, ts.URL, bobCookie)
	seenBefore, found := findDirectChatSeen(bobsChats.Items, 1)
	if !found {
		log.Fatalf("expected bob to have a direct chat with alice")
	}
	if seenBefore {
		log.Fatalf("expected direct chat seen=false before mark-read")
	}

	markRead(client, ts.URL, bobCookie, directMsg.ChatID, directMsg.ID)

	bobsChatsAfter := listChats(client, ts.URL, bobCookie)
	seenAfter, found := findDirectChatSeen(bobsChatsAfter.Items, 1)
	if !found {
		log.Fatalf("expected bob to still have a direct chat with alice after mark-read")
	}
	if !seenAfter {
		log.Fatalf("expected direct chat seen=true after mark-read")
	}

	// ---- Group chat flow + edge case: leave removes access ----
	groupID := createGroup(client, ts.URL, aliceCookie, "auto")
	joinGroup(client, ts.URL, bobCookie, groupID)

	groupMsg := sendGroupMessage(client, ts.URL, bobCookie, groupID, "hello group", http.StatusCreated)
	if groupMsg.ChatID <= 0 {
		log.Fatalf("expected valid group chat_id from message")
	}

	getMessagesExpect(client, ts.URL, bobCookie, groupMsg.ChatID, http.StatusOK)
	leaveGroup(client, ts.URL, bobCookie, groupID)
	getMessagesExpect(client, ts.URL, bobCookie, groupMsg.ChatID, http.StatusForbidden)
	sendGroupMessage(client, ts.URL, bobCookie, groupID, "should fail after leave", http.StatusForbidden)

	// extra edge: non-member cannot send to group chat
	sendGroupMessage(client, ts.URL, carolCookie, groupID, "carol not member", http.StatusForbidden)

	fmt.Println("chat integration check passed (direct + read-state + group access edge cases)")
}
