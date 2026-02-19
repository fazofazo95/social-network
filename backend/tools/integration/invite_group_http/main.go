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

func createGroup(client *http.Client, baseURL string, ownerCookie *http.Cookie, name, joinMode string) (int, error) {
	payload := map[string]string{
		"name":        name,
		"description": "invite integration group",
		"visibility":  "public",
		"join_mode":   joinMode,
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

func invite(client *http.Client, baseURL string, inviterCookie *http.Cookie, groupID, targetUserID int) (int, string, error) {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/invite/%d", baseURL, groupID, targetUserID), nil)
	req.AddCookie(inviterCookie)
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(out), nil
}

func verifyRequestedInviteState(groupID, targetUserID, inviterID, expectedMemberCount int) error {
	var memberStatus, role string
	if err := database.DB.QueryRow(`
		SELECT status, role
		FROM group_members
		WHERE group_id = ? AND user_id = ?
	`, groupID, targetUserID).Scan(&memberStatus, &role); err != nil {
		return fmt.Errorf("verify group_members row: %w", err)
	}
	if memberStatus != "requested" {
		return fmt.Errorf("expected group_members.status=requested got %s", memberStatus)
	}
	if role != "member" {
		return fmt.Errorf("expected group_members.role=member got %s", role)
	}

	var reqStatus, reqType string
	var actionByID int
	if err := database.DB.QueryRow(`
		SELECT status, request_type, action_by_id
		FROM group_join_requests
		WHERE group_id = ? AND user_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, groupID, targetUserID).Scan(&reqStatus, &reqType, &actionByID); err != nil {
		return fmt.Errorf("verify group_join_requests row: %w", err)
	}
	if reqStatus != "invite" {
		return fmt.Errorf("expected group_join_requests.status=invite got %s", reqStatus)
	}
	if reqType != "invite" {
		return fmt.Errorf("expected group_join_requests.request_type=invite got %s", reqType)
	}
	if actionByID != inviterID {
		return fmt.Errorf("expected action_by_id=%d got %d", inviterID, actionByID)
	}

	var membersCount int
	if err := database.DB.QueryRow(`SELECT group_members FROM groups WHERE id = ?`, groupID).Scan(&membersCount); err != nil {
		return fmt.Errorf("verify groups.group_members: %w", err)
	}
	if membersCount != expectedMemberCount {
		return fmt.Errorf("expected group_members count to stay %d on invite pending, got %d", expectedMemberCount, membersCount)
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

	ownerGroupID, err := createGroup(client, ts.URL, aliceCookie, fmt.Sprintf("Invite Owner Request %d", time.Now().UnixNano()), "request")
	if err != nil {
		log.Fatalf("create owner group: %v", err)
	}
	statusCode, body, err := invite(client, ts.URL, aliceCookie, ownerGroupID, 3)
	if err != nil {
		log.Fatalf("owner invite call: %v", err)
	}
	fmt.Printf("owner invite -> %d %s\n", statusCode, body)
	if statusCode != http.StatusOK {
		log.Fatalf("owner invite expected 200 got %d", statusCode)
	}
	if err := verifyRequestedInviteState(ownerGroupID, 3, 1, 1); err != nil {
		log.Fatalf("owner invite verification failed: %v", err)
	}

	moderatorGroupID, err := createGroup(client, ts.URL, aliceCookie, fmt.Sprintf("Invite Moderator Request %d", time.Now().UnixNano()), "request")
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
		log.Fatalf("update group_members after moderator seed: %v", err)
	}

	statusCode, body, err = invite(client, ts.URL, bobCookie, moderatorGroupID, 4)
	if err != nil {
		log.Fatalf("moderator invite call: %v", err)
	}
	fmt.Printf("moderator invite -> %d %s\n", statusCode, body)
	if statusCode != http.StatusOK {
		log.Fatalf("moderator invite expected 200 got %d", statusCode)
	}
	if err := verifyRequestedInviteState(moderatorGroupID, 4, 2, 2); err != nil {
		log.Fatalf("moderator invite verification failed: %v", err)
	}

	autoGroupID, err := createGroup(client, ts.URL, aliceCookie, fmt.Sprintf("Invite Owner Auto %d", time.Now().UnixNano()), "auto")
	if err != nil {
		log.Fatalf("create auto group: %v", err)
	}
	statusCode, body, err = invite(client, ts.URL, aliceCookie, autoGroupID, 5)
	if err != nil {
		log.Fatalf("auto mode invite call: %v", err)
	}
	fmt.Printf("auto mode invite -> %d %s\n", statusCode, body)
	if statusCode != http.StatusOK {
		log.Fatalf("auto mode invite expected 200 got %d", statusCode)
	}
	if err := verifyRequestedInviteState(autoGroupID, 5, 1, 1); err != nil {
		log.Fatalf("auto mode invite verification failed: %v", err)
	}

	inviteModeGroupID, err := createGroup(client, ts.URL, aliceCookie, fmt.Sprintf("Invite Owner InviteMode %d", time.Now().UnixNano()), "invite")
	if err != nil {
		log.Fatalf("create invite-mode group: %v", err)
	}
	statusCode, body, err = invite(client, ts.URL, aliceCookie, inviteModeGroupID, 6)
	if err != nil {
		log.Fatalf("invite mode invite call: %v", err)
	}
	fmt.Printf("invite mode invite -> %d %s\n", statusCode, body)
	if statusCode != http.StatusOK {
		log.Fatalf("invite mode invite expected 200 got %d", statusCode)
	}
	if err := verifyRequestedInviteState(inviteModeGroupID, 6, 1, 1); err != nil {
		log.Fatalf("invite mode invite verification failed: %v", err)
	}

	fmt.Println("invite group integration check passed")
}
