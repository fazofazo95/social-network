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
	ownerCookie, err := login(client, ts.URL, "alice@example.com", "Password123!")
	if err != nil {
		log.Fatalf("owner login error: %v", err)
	}

	otherCookie, err := login(client, ts.URL, "bob@example.com", "Password123!")
	if err != nil {
		log.Fatalf("other user login error: %v", err)
	}

	createPayload := map[string]string{
		"name":        fmt.Sprintf("Delete Group Forbidden Demo %d", time.Now().UnixNano()),
		"description": "Group to test non-owner delete",
		"visibility":  "public",
		"join_mode":   "request",
	}
	createBody, _ := json.Marshal(createPayload)

	createReq, _ := http.NewRequest(http.MethodPost, ts.URL+"/api/groups", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(ownerCookie)

	createResp, err := client.Do(createReq)
	if err != nil {
		log.Fatalf("create group error: %v", err)
	}
	defer createResp.Body.Close()

	createOut, _ := io.ReadAll(createResp.Body)
	if createResp.StatusCode != http.StatusCreated {
		log.Fatalf("create group failed: %d %s", createResp.StatusCode, string(createOut))
	}

	var createEnvelope apiResponse
	if err := json.Unmarshal(createOut, &createEnvelope); err != nil {
		log.Fatalf("decode create response envelope: %v", err)
	}

	var created createGroupData
	if err := json.Unmarshal(createEnvelope.Data, &created); err != nil {
		log.Fatalf("decode create response data: %v", err)
	}
	if created.ID <= 0 {
		log.Fatalf("invalid group id returned: %d", created.ID)
	}

	groupID := created.ID

	forbiddenReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/groups/%d", ts.URL, groupID), nil)
	forbiddenReq.AddCookie(otherCookie)

	forbiddenResp, err := client.Do(forbiddenReq)
	if err != nil {
		log.Fatalf("non-owner delete call error: %v", err)
	}
	defer forbiddenResp.Body.Close()

	forbiddenOut, _ := io.ReadAll(forbiddenResp.Body)
	fmt.Printf("DELETE(non-owner) /api/groups/%d -> %d\n%s\n", groupID, forbiddenResp.StatusCode, string(forbiddenOut))

	if forbiddenResp.StatusCode != http.StatusForbidden {
		log.Fatalf("expected 403 for non-owner delete, got %d", forbiddenResp.StatusCode)
	}

	var groupsCount int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM groups WHERE id = ?`, groupID).Scan(&groupsCount); err != nil {
		log.Fatalf("count groups after forbidden delete: %v", err)
	}
	fmt.Printf("group rows after forbidden delete => %d\n", groupsCount)
	if groupsCount != 1 {
		log.Fatalf("expected group to remain after forbidden delete")
	}

	cleanupReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/groups/%d", ts.URL, groupID), nil)
	cleanupReq.AddCookie(ownerCookie)
	cleanupResp, err := client.Do(cleanupReq)
	if err != nil {
		log.Fatalf("cleanup delete call error: %v", err)
	}
	defer cleanupResp.Body.Close()
	if cleanupResp.StatusCode != http.StatusOK {
		cleanupOut, _ := io.ReadAll(cleanupResp.Body)
		log.Fatalf("cleanup delete failed: %d %s", cleanupResp.StatusCode, string(cleanupOut))
	}

	fmt.Println("non-owner delete integration check passed")
}
