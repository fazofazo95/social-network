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
		"name":        fmt.Sprintf("Remove Pending %d", time.Now().UnixNano()),
		"description": "remove pending integration",
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
	daveCookie, err := login(client, ts.URL, "dave@example.com", "Password123!")
	if err != nil {
		log.Fatalf("dave login: %v", err)
	}

	groupID, err := createGroup(client, ts.URL, aliceCookie)
	if err != nil {
		log.Fatalf("create group: %v", err)
	}

	if _, err := database.DB.Exec(`
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES (?, 2, 'moderator', 'active')
	`, groupID); err != nil {
		log.Fatalf("seed moderator: %v", err)
	}
	if _, err := database.DB.Exec(`UPDATE groups SET group_members = group_members + 1 WHERE id = ?`, groupID); err != nil {
		log.Fatalf("update count after moderator seed: %v", err)
	}

	inviteReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/invite/%d", ts.URL, groupID, 4), nil)
	inviteReq.AddCookie(aliceCookie)
	inviteResp, err := client.Do(inviteReq)
	if err != nil {
		log.Fatalf("invite user call: %v", err)
	}
	inviteOut, _ := io.ReadAll(inviteResp.Body)
	inviteResp.Body.Close()
	if inviteResp.StatusCode != http.StatusOK {
		log.Fatalf("invite user failed: %d %s", inviteResp.StatusCode, string(inviteOut))
	}

	removeInviteReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/groups/%d/invites/%d", ts.URL, groupID, 4), nil)
	removeInviteReq.AddCookie(bobCookie)
	removeInviteResp, err := client.Do(removeInviteReq)
	if err != nil {
		log.Fatalf("remove invite call: %v", err)
	}
	removeInviteOut, _ := io.ReadAll(removeInviteResp.Body)
	removeInviteResp.Body.Close()
	fmt.Printf("remove invite (moderator) -> %d %s\n", removeInviteResp.StatusCode, string(removeInviteOut))
	if removeInviteResp.StatusCode != http.StatusOK {
		log.Fatalf("expected remove invite 200 got %d", removeInviteResp.StatusCode)
	}

	var inviteMemberRows, inviteRequestRows int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM group_members WHERE group_id = ? AND user_id = ?`, groupID, 4).Scan(&inviteMemberRows); err != nil {
		log.Fatalf("check invite member rows: %v", err)
	}
	if err := database.DB.QueryRow(`
		SELECT COUNT(*) FROM group_join_requests
		WHERE group_id = ? AND user_id = ? AND request_type = 'invite' AND status = 'invite'
	`, groupID, 4).Scan(&inviteRequestRows); err != nil {
		log.Fatalf("check invite request rows: %v", err)
	}
	if inviteMemberRows != 0 || inviteRequestRows != 0 {
		log.Fatalf("expected invite rows removed; memberRows=%d requestRows=%d", inviteMemberRows, inviteRequestRows)
	}

	joinReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/join", ts.URL, groupID), nil)
	joinReq.AddCookie(carolCookie)
	joinResp, err := client.Do(joinReq)
	if err != nil {
		log.Fatalf("join request call: %v", err)
	}
	joinOut, _ := io.ReadAll(joinResp.Body)
	joinResp.Body.Close()
	if joinResp.StatusCode != http.StatusOK {
		log.Fatalf("join request failed: %d %s", joinResp.StatusCode, string(joinOut))
	}

	removeOwnReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/groups/%d/requests/me", ts.URL, groupID), nil)
	removeOwnReq.AddCookie(carolCookie)
	removeOwnResp, err := client.Do(removeOwnReq)
	if err != nil {
		log.Fatalf("remove own request call: %v", err)
	}
	removeOwnOut, _ := io.ReadAll(removeOwnResp.Body)
	removeOwnResp.Body.Close()
	fmt.Printf("remove own request -> %d %s\n", removeOwnResp.StatusCode, string(removeOwnOut))
	if removeOwnResp.StatusCode != http.StatusOK {
		log.Fatalf("expected remove own request 200 got %d", removeOwnResp.StatusCode)
	}

	var reqMemberRows, reqRequestRows int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM group_members WHERE group_id = ? AND user_id = ?`, groupID, 3).Scan(&reqMemberRows); err != nil {
		log.Fatalf("check request member rows: %v", err)
	}
	if err := database.DB.QueryRow(`
		SELECT COUNT(*) FROM group_join_requests
		WHERE group_id = ? AND user_id = ? AND request_type = 'request' AND status = 'request'
	`, groupID, 3).Scan(&reqRequestRows); err != nil {
		log.Fatalf("check request rows: %v", err)
	}
	if reqMemberRows != 0 || reqRequestRows != 0 {
		log.Fatalf("expected request rows removed; memberRows=%d requestRows=%d", reqMemberRows, reqRequestRows)
	}

	nonModRemoveInviteReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/groups/%d/invites/%d", ts.URL, groupID, 2), nil)
	nonModRemoveInviteReq.AddCookie(daveCookie)
	nonModRemoveInviteResp, err := client.Do(nonModRemoveInviteReq)
	if err != nil {
		log.Fatalf("non-mod remove invite call: %v", err)
	}
	nonModRemoveInviteOut, _ := io.ReadAll(nonModRemoveInviteResp.Body)
	nonModRemoveInviteResp.Body.Close()
	fmt.Printf("remove invite (non-mod) -> %d %s\n", nonModRemoveInviteResp.StatusCode, string(nonModRemoveInviteOut))
	if nonModRemoveInviteResp.StatusCode != http.StatusForbidden {
		log.Fatalf("expected non-mod remove invite 403 got %d", nonModRemoveInviteResp.StatusCode)
	}

	fmt.Println("remove pending invite/request integration check passed")
}
