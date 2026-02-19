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

type apiResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

type createGroupData struct {
	ID int `json:"id"`
}

func login(client *http.Client, baseURL, email, password string) (*http.Cookie, error) {
	payload := map[string]string{"email": email, "password": password}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/login", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		out, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("login failed: %d %s", resp.StatusCode, string(out))
	}
	for _, c := range resp.Cookies() {
		if c.Name == "session_id" {
			return c, nil
		}
	}
	return nil, fmt.Errorf("missing session_id cookie")
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
	handlers.PostRoutes(mux, database.DB)
	handlers.CommentRoutes(mux, database.DB)
	handlers.FeedRoutes(mux)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := ts.Client()
	cookie, err := login(client, ts.URL, "alice@example.com", "Password123!")
	if err != nil {
		log.Fatalf("login error: %v", err)
	}

	createPayload := map[string]string{
		"name":        fmt.Sprintf("Delete Group Demo %d", time.Now().UnixNano()),
		"description": "Group to test delete endpoint",
		"visibility":  "public",
		"join_mode":   "request",
	}
	createBody, _ := json.Marshal(createPayload)

	createReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/groups", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(cookie)

	createResp, err := client.Do(createReq)
	if err != nil {
		log.Fatalf("create group error: %v", err)
	}
	defer createResp.Body.Close()

	createOut, _ := io.ReadAll(createResp.Body)
	if createResp.StatusCode != http.StatusCreated {
		log.Fatalf("create group failed: %d %s", createResp.StatusCode, string(createOut))
	}

	var respEnvelope apiResponse
	if err := json.Unmarshal(createOut, &respEnvelope); err != nil {
		log.Fatalf("decode create response envelope: %v", err)
	}

	var created createGroupData
	if err := json.Unmarshal(respEnvelope.Data, &created); err != nil {
		log.Fatalf("decode create response data: %v", err)
	}
	if created.ID <= 0 {
		log.Fatalf("invalid group id returned: %d", created.ID)
	}

	groupID := created.ID

	if _, err := database.DB.Exec(`
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES (?, 2, 'member', 'active')
	`, groupID); err != nil {
		log.Fatalf("seed group member row: %v", err)
	}

	if _, err := database.DB.Exec(`
		INSERT INTO group_join_requests (group_id, user_id, request_type, status)
		VALUES (?, 2, 'request', 'request')
	`, groupID); err != nil {
		log.Fatalf("seed group request row: %v", err)
	}

	deleteReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/groups/%d", ts.URL, groupID), nil)
	deleteReq.AddCookie(cookie)

	deleteResp, err := client.Do(deleteReq)
	if err != nil {
		log.Fatalf("delete group error: %v", err)
	}
	defer deleteResp.Body.Close()

	deleteOut, _ := io.ReadAll(deleteResp.Body)
	if deleteResp.StatusCode != http.StatusOK {
		log.Fatalf("delete group failed: %d %s", deleteResp.StatusCode, string(deleteOut))
	}

	var groupsCount, membersCount, requestsCount int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM groups WHERE id = ?`, groupID).Scan(&groupsCount); err != nil {
		log.Fatalf("count groups: %v", err)
	}
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM group_members WHERE group_id = ?`, groupID).Scan(&membersCount); err != nil {
		log.Fatalf("count group_members: %v", err)
	}
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM group_join_requests WHERE group_id = ?`, groupID).Scan(&requestsCount); err != nil {
		log.Fatalf("count group_join_requests: %v", err)
	}

	fmt.Printf("DELETE /api/groups/%d -> %d\n%s\n", groupID, deleteResp.StatusCode, string(deleteOut))
	fmt.Printf("rows after delete => groups:%d members:%d requests:%d\n", groupsCount, membersCount, requestsCount)

	if groupsCount != 0 || membersCount != 0 || requestsCount != 0 {
		log.Fatalf("delete verification failed: expected all zero counts")
	}

	fmt.Println("delete group integration check passed")
}
