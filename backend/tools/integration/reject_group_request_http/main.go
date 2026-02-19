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

func createGroup(client *http.Client, baseURL string, ownerCookie *http.Cookie) (int, error) {
	payload := map[string]string{
		"name":        fmt.Sprintf("Reject Request %d", time.Now().UnixNano()),
		"description": "reject request integration group",
		"visibility":  "public",
		"join_mode":   "request",
	}
	b, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/groups", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(ownerCookie)
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated {
		return 0, fmt.Errorf("create group failed: %d %s", resp.StatusCode, string(out))
	}
	var env apiEnvelope
	if err := json.Unmarshal(out, &env); err != nil {
		return 0, err
	}
	var data createGroupData
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return 0, err
	}
	if data.ID <= 0 {
		return 0, fmt.Errorf("invalid group id")
	}
	return data.ID, nil
}

func requestJoin(client *http.Client, baseURL string, userCookie *http.Cookie, groupID int) (int, string, error) {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/join", baseURL, groupID), nil)
	req.AddCookie(userCookie)
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(out), nil
}

func rejectJoin(client *http.Client, baseURL string, approverCookie *http.Cookie, groupID, requesterID int) (int, string, error) {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/requests/%d/reject", baseURL, groupID, requesterID), nil)
	req.AddCookie(approverCookie)
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(out), nil
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
		log.Fatalf("owner login: %v", err)
	}
	requesterCookie, err := login(client, ts.URL, "bob@example.com", "Password123!")
	if err != nil {
		log.Fatalf("requester login: %v", err)
	}

	groupID, err := createGroup(client, ts.URL, ownerCookie)
	if err != nil {
		log.Fatalf("create group: %v", err)
	}

	statusCode, body, err := requestJoin(client, ts.URL, requesterCookie, groupID)
	if err != nil {
		log.Fatalf("initial request join call: %v", err)
	}
	fmt.Printf("initial join request -> %d %s\n", statusCode, body)
	if statusCode != http.StatusOK {
		log.Fatalf("expected initial join request status 200 got %d", statusCode)
	}

	statusCode, body, err = rejectJoin(client, ts.URL, ownerCookie, groupID, 2)
	if err != nil {
		log.Fatalf("reject call: %v", err)
	}
	fmt.Printf("reject request -> %d %s\n", statusCode, body)
	if statusCode != http.StatusOK {
		log.Fatalf("expected reject status 200 got %d", statusCode)
	}

	var memberRows int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM group_members WHERE group_id = ? AND user_id = 2`, groupID).Scan(&memberRows); err != nil {
		log.Fatalf("count group_members rows: %v", err)
	}
	if memberRows != 0 {
		log.Fatalf("expected no group_members row after reject, got %d", memberRows)
	}

	var reqStatus, reqType string
	var actionByID int
	if err := database.DB.QueryRow(`
		SELECT status, request_type, action_by_id
		FROM group_join_requests
		WHERE group_id = ? AND user_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, groupID, 2).Scan(&reqStatus, &reqType, &actionByID); err != nil {
		log.Fatalf("read latest group_join_requests row: %v", err)
	}
	if reqStatus != "rejected" {
		log.Fatalf("expected latest request status rejected, got %s", reqStatus)
	}
	if reqType != "request" {
		log.Fatalf("expected latest request_type request, got %s", reqType)
	}
	if actionByID != 1 {
		log.Fatalf("expected action_by_id=1 got %d", actionByID)
	}

	statusCode, body, err = requestJoin(client, ts.URL, requesterCookie, groupID)
	if err != nil {
		log.Fatalf("re-apply join call: %v", err)
	}
	fmt.Printf("re-apply join request -> %d %s\n", statusCode, body)
	if statusCode != http.StatusOK {
		log.Fatalf("expected re-apply join status 200 got %d", statusCode)
	}

	var reapplyMemberStatus string
	if err := database.DB.QueryRow(`SELECT status FROM group_members WHERE group_id = ? AND user_id = 2`, groupID).Scan(&reapplyMemberStatus); err != nil {
		log.Fatalf("read reapply member row: %v", err)
	}
	if reapplyMemberStatus != "requested" {
		log.Fatalf("expected re-apply member status requested, got %s", reapplyMemberStatus)
	}

	fmt.Println("reject group request integration check passed")
}
