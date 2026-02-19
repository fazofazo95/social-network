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

func createGroup(client *http.Client, baseURL string, ownerCookie *http.Cookie, name string) (int, error) {
	payload := map[string]string{
		"name":        name,
		"description": "accept request integration group",
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

func requestJoin(client *http.Client, baseURL string, cookie *http.Cookie, groupID int) error {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/join", baseURL, groupID), nil)
	req.AddCookie(cookie)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request join failed: %d %s", resp.StatusCode, string(out))
	}
	return nil
}

func approveRequest(client *http.Client, baseURL string, approverCookie *http.Cookie, groupID, requesterID int) (int, string, error) {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/requests/%d/accept", baseURL, groupID, requesterID), nil)
	req.AddCookie(approverCookie)
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(out), nil
}

func verifyApprovedState(groupID, requesterID, approverID, expectedGroupMembers int) error {
	var memberStatus, role, joinedAt string
	if err := database.DB.QueryRow(`
		SELECT status, role, COALESCE(datetime(joined_at), '')
		FROM group_members
		WHERE group_id = ? AND user_id = ?
	`, groupID, requesterID).Scan(&memberStatus, &role, &joinedAt); err != nil {
		return fmt.Errorf("verify group_members row: %w", err)
	}

	if memberStatus != "active" {
		return fmt.Errorf("expected group_members.status=active got %s", memberStatus)
	}
	if role != "member" {
		return fmt.Errorf("expected group_members.role=member got %s", role)
	}
	if joinedAt == "" {
		return fmt.Errorf("expected non-empty joined_at")
	}

	var membersCount int
	if err := database.DB.QueryRow(`SELECT group_members FROM groups WHERE id = ?`, groupID).Scan(&membersCount); err != nil {
		return fmt.Errorf("verify groups.group_members: %w", err)
	}
	if membersCount != expectedGroupMembers {
		return fmt.Errorf("expected groups.group_members=%d got %d", expectedGroupMembers, membersCount)
	}

	var reqStatus, reqType string
	var actionByID int
	if err := database.DB.QueryRow(`
		SELECT status, request_type, action_by_id
		FROM group_join_requests
		WHERE group_id = ? AND user_id = ?
	`, groupID, requesterID).Scan(&reqStatus, &reqType, &actionByID); err != nil {
		return fmt.Errorf("verify group_join_requests row: %w", err)
	}
	if reqStatus != "approved" {
		return fmt.Errorf("expected group_join_requests.status=approved got %s", reqStatus)
	}
	if reqType != "request" {
		return fmt.Errorf("expected group_join_requests.request_type=request got %s", reqType)
	}
	if actionByID != approverID {
		return fmt.Errorf("expected action_by_id=%d got %d", approverID, actionByID)
	}

	return nil
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
	aliceCookie, err := login(client, ts.URL, "alice@example.com", "Password123!")
	if err != nil {
		log.Fatalf("alice login: %v", err)
	}
	bobCookie, err := login(client, ts.URL, "bob@example.com", "Password123!")
	if err != nil {
		log.Fatalf("bob login: %v", err)
	}
	carolCookie, err := login(client, ts.URL, "carol@example.com", "Password123!")
	if err != nil {
		log.Fatalf("carol login: %v", err)
	}

	ownerGroupID, err := createGroup(client, ts.URL, aliceCookie, fmt.Sprintf("Accept Request Owner %d", time.Now().UnixNano()))
	if err != nil {
		log.Fatalf("create owner group: %v", err)
	}
	if err := requestJoin(client, ts.URL, bobCookie, ownerGroupID); err != nil {
		log.Fatalf("owner scenario request join: %v", err)
	}
	statusCode, body, err := approveRequest(client, ts.URL, aliceCookie, ownerGroupID, 2)
	if err != nil {
		log.Fatalf("owner scenario approve request call: %v", err)
	}
	fmt.Printf("owner approve -> %d %s\n", statusCode, body)
	if statusCode != http.StatusOK {
		log.Fatalf("owner scenario expected 200 got %d", statusCode)
	}
	if err := verifyApprovedState(ownerGroupID, 2, 1, 2); err != nil {
		log.Fatalf("owner scenario verification failed: %v", err)
	}

	moderatorGroupID, err := createGroup(client, ts.URL, aliceCookie, fmt.Sprintf("Accept Request Moderator %d", time.Now().UnixNano()))
	if err != nil {
		log.Fatalf("create moderator group: %v", err)
	}
	if _, err := database.DB.Exec(`
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES (?, 2, 'moderator', 'active')
	`, moderatorGroupID); err != nil {
		log.Fatalf("seed moderator membership: %v", err)
	}
	if _, err := database.DB.Exec(`
		UPDATE groups SET group_members = group_members + 1 WHERE id = ?
	`, moderatorGroupID); err != nil {
		log.Fatalf("update group_members after seeding moderator: %v", err)
	}

	if err := requestJoin(client, ts.URL, carolCookie, moderatorGroupID); err != nil {
		log.Fatalf("moderator scenario request join: %v", err)
	}
	statusCode, body, err = approveRequest(client, ts.URL, bobCookie, moderatorGroupID, 3)
	if err != nil {
		log.Fatalf("moderator scenario approve request call: %v", err)
	}
	fmt.Printf("moderator approve -> %d %s\n", statusCode, body)
	if statusCode != http.StatusOK {
		log.Fatalf("moderator scenario expected 200 got %d", statusCode)
	}
	if err := verifyApprovedState(moderatorGroupID, 3, 2, 3); err != nil {
		log.Fatalf("moderator scenario verification failed: %v", err)
	}

	fmt.Println("accept group request integration check passed")
}
