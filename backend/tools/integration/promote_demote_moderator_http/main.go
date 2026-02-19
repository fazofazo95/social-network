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
		"description": "promote/demote integration group",
		"visibility":  "public",
		"join_mode":   "request_and_invite",
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
	return data.ID, nil
}

func addActiveMember(groupID, userID int) error {
	if _, err := database.DB.Exec(`
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES (?, ?, 'member', 'active')
	`, groupID, userID); err != nil {
		return err
	}
	if _, err := database.DB.Exec(`UPDATE groups SET group_members = group_members + 1 WHERE id = ?`, groupID); err != nil {
		return err
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

	groupID, err := createGroup(client, ts.URL, aliceCookie, fmt.Sprintf("Promote Demote %d", time.Now().UnixNano()))
	if err != nil {
		log.Fatalf("create group: %v", err)
	}
	if err := addActiveMember(groupID, 2); err != nil {
		log.Fatalf("seed member bob: %v", err)
	}

	nonOwnerPromoteReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/members/%d/promote", ts.URL, groupID, 1), nil)
	nonOwnerPromoteReq.AddCookie(bobCookie)
	nonOwnerPromoteResp, err := client.Do(nonOwnerPromoteReq)
	if err != nil {
		log.Fatalf("non-owner promote call: %v", err)
	}
	nonOwnerPromoteOut, _ := io.ReadAll(nonOwnerPromoteResp.Body)
	nonOwnerPromoteResp.Body.Close()
	fmt.Printf("non-owner promote attempt -> %d %s\n", nonOwnerPromoteResp.StatusCode, string(nonOwnerPromoteOut))
	if nonOwnerPromoteResp.StatusCode != http.StatusForbidden {
		log.Fatalf("expected non-owner promote 403 got %d", nonOwnerPromoteResp.StatusCode)
	}

	promoteReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/members/%d/promote", ts.URL, groupID, 2), nil)
	promoteReq.AddCookie(aliceCookie)
	promoteResp, err := client.Do(promoteReq)
	if err != nil {
		log.Fatalf("promote call: %v", err)
	}
	promoteOut, _ := io.ReadAll(promoteResp.Body)
	promoteResp.Body.Close()
	fmt.Printf("owner promotes member -> %d %s\n", promoteResp.StatusCode, string(promoteOut))
	if promoteResp.StatusCode != http.StatusOK {
		log.Fatalf("expected promote 200 got %d", promoteResp.StatusCode)
	}
	var role string
	if err := database.DB.QueryRow(`SELECT role FROM group_members WHERE group_id = ? AND user_id = ?`, groupID, 2).Scan(&role); err != nil {
		log.Fatalf("verify role after promote: %v", err)
	}
	if role != "moderator" {
		log.Fatalf("expected role moderator after promote, got %s", role)
	}

	alreadyModeratorPromoteReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/members/%d/promote", ts.URL, groupID, 2), nil)
	alreadyModeratorPromoteReq.AddCookie(aliceCookie)
	alreadyModeratorPromoteResp, err := client.Do(alreadyModeratorPromoteReq)
	if err != nil {
		log.Fatalf("promote already-moderator call: %v", err)
	}
	alreadyModeratorPromoteOut, _ := io.ReadAll(alreadyModeratorPromoteResp.Body)
	alreadyModeratorPromoteResp.Body.Close()
	fmt.Printf("owner promotes already-moderator -> %d %s\n", alreadyModeratorPromoteResp.StatusCode, string(alreadyModeratorPromoteOut))
	if alreadyModeratorPromoteResp.StatusCode != http.StatusConflict {
		log.Fatalf("expected promote already-moderator 409 got %d", alreadyModeratorPromoteResp.StatusCode)
	}

	demoteReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/members/%d/demote", ts.URL, groupID, 2), nil)
	demoteReq.AddCookie(aliceCookie)
	demoteResp, err := client.Do(demoteReq)
	if err != nil {
		log.Fatalf("demote call: %v", err)
	}
	demoteOut, _ := io.ReadAll(demoteResp.Body)
	demoteResp.Body.Close()
	fmt.Printf("owner demotes moderator -> %d %s\n", demoteResp.StatusCode, string(demoteOut))
	if demoteResp.StatusCode != http.StatusOK {
		log.Fatalf("expected demote 200 got %d", demoteResp.StatusCode)
	}
	if err := database.DB.QueryRow(`SELECT role FROM group_members WHERE group_id = ? AND user_id = ?`, groupID, 2).Scan(&role); err != nil {
		log.Fatalf("verify role after demote: %v", err)
	}
	if role != "member" {
		log.Fatalf("expected role member after demote, got %s", role)
	}

	alreadyMemberDemoteReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/members/%d/demote", ts.URL, groupID, 2), nil)
	alreadyMemberDemoteReq.AddCookie(aliceCookie)
	alreadyMemberDemoteResp, err := client.Do(alreadyMemberDemoteReq)
	if err != nil {
		log.Fatalf("demote already-member call: %v", err)
	}
	alreadyMemberDemoteOut, _ := io.ReadAll(alreadyMemberDemoteResp.Body)
	alreadyMemberDemoteResp.Body.Close()
	fmt.Printf("owner demotes already-member -> %d %s\n", alreadyMemberDemoteResp.StatusCode, string(alreadyMemberDemoteOut))
	if alreadyMemberDemoteResp.StatusCode != http.StatusConflict {
		log.Fatalf("expected demote already-member 409 got %d", alreadyMemberDemoteResp.StatusCode)
	}

	fmt.Println("promote/demote moderator integration check passed")
}
