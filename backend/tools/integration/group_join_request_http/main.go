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
	ID       int    `json:"id"`
	JoinMode string `json:"join_mode"`
}

type joinData struct {
	GroupID          int    `json:"group_id"`
	UserID           int    `json:"user_id"`
	MembershipStatus string `json:"membership_status"`
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

func createGroup(client *http.Client, baseURL string, ownerCookie *http.Cookie, visibility, joinMode string) (int, error) {
	payload := map[string]string{
		"name":        fmt.Sprintf("Join Mode %s %d", joinMode, time.Now().UnixNano()),
		"description": "integration test group",
		"visibility":  visibility,
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
	joinerCookie, err := login(client, ts.URL, "bob@example.com", "Password123!")
	if err != nil {
		log.Fatalf("joiner login: %v", err)
	}

	autoGroupID, err := createGroup(client, ts.URL, ownerCookie, "public", "auto")
	if err != nil {
		log.Fatalf("create auto group: %v", err)
	}
	requestGroupID, err := createGroup(client, ts.URL, ownerCookie, "public", "request")
	if err != nil {
		log.Fatalf("create request group: %v", err)
	}
	requestInviteGroupID, err := createGroup(client, ts.URL, ownerCookie, "public", "request_and_invite")
	if err != nil {
		log.Fatalf("create request_and_invite group: %v", err)
	}
	inviteOnlyGroupID, err := createGroup(client, ts.URL, ownerCookie, "public", "invite")
	if err != nil {
		log.Fatalf("create invite group: %v", err)
	}
	privateGroupID, err := createGroup(client, ts.URL, ownerCookie, "private", "request")
	if err != nil {
		log.Fatalf("create private group: %v", err)
	}

	type scenario struct {
		name           string
		groupID        int
		expectedCode   int
		expectedStatus string
	}

	scenarios := []scenario{
		{name: "public+auto", groupID: autoGroupID, expectedCode: http.StatusOK, expectedStatus: "active"},
		{name: "public+request", groupID: requestGroupID, expectedCode: http.StatusOK, expectedStatus: "requested"},
		{name: "public+request_and_invite", groupID: requestInviteGroupID, expectedCode: http.StatusOK, expectedStatus: "requested"},
		{name: "public+invite blocked", groupID: inviteOnlyGroupID, expectedCode: http.StatusForbidden, expectedStatus: ""},
		{name: "private blocked", groupID: privateGroupID, expectedCode: http.StatusForbidden, expectedStatus: ""},
	}

	for _, sc := range scenarios {
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/join", ts.URL, sc.groupID), nil)
		req.AddCookie(joinerCookie)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("%s call failed: %v", sc.name, err)
		}
		out, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		fmt.Printf("%s -> %d %s\n", sc.name, resp.StatusCode, string(out))

		if resp.StatusCode != sc.expectedCode {
			log.Fatalf("%s expected status %d got %d", sc.name, sc.expectedCode, resp.StatusCode)
		}

		if sc.expectedCode == http.StatusOK {
			var env apiEnvelope
			if err := json.Unmarshal(out, &env); err != nil {
				log.Fatalf("%s decode envelope: %v", sc.name, err)
			}
			var data joinData
			if err := json.Unmarshal(env.Data, &data); err != nil {
				log.Fatalf("%s decode data: %v", sc.name, err)
			}
			if data.MembershipStatus != sc.expectedStatus {
				log.Fatalf("%s expected membership_status %q got %q", sc.name, sc.expectedStatus, data.MembershipStatus)
			}
		}
	}

	var autoMemberStatus string
	if err := database.DB.QueryRow(`SELECT status FROM group_members WHERE group_id = ? AND user_id = 2`, autoGroupID).Scan(&autoMemberStatus); err != nil {
		log.Fatalf("verify auto member status: %v", err)
	}
	if autoMemberStatus != "active" {
		log.Fatalf("auto mode expected active member, got %s", autoMemberStatus)
	}

	var autoMemberCount int
	if err := database.DB.QueryRow(`SELECT group_members FROM groups WHERE id = ?`, autoGroupID).Scan(&autoMemberCount); err != nil {
		log.Fatalf("verify auto group_members count: %v", err)
	}
	if autoMemberCount != 2 {
		log.Fatalf("auto mode expected group_members=2, got %d", autoMemberCount)
	}

	var reqMemberStatus string
	if err := database.DB.QueryRow(`SELECT status FROM group_members WHERE group_id = ? AND user_id = 2`, requestGroupID).Scan(&reqMemberStatus); err != nil {
		log.Fatalf("verify request member status: %v", err)
	}
	if reqMemberStatus != "requested" {
		log.Fatalf("request mode expected requested member, got %s", reqMemberStatus)
	}

	var reqJoinRequestCount int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM group_join_requests WHERE group_id = ? AND user_id = 2 AND status = 'request'`, requestGroupID).Scan(&reqJoinRequestCount); err != nil {
		log.Fatalf("verify request join request row: %v", err)
	}
	if reqJoinRequestCount != 1 {
		log.Fatalf("request mode expected one join request row, got %d", reqJoinRequestCount)
	}

	var reqInviteMemberStatus string
	if err := database.DB.QueryRow(`SELECT status FROM group_members WHERE group_id = ? AND user_id = 2`, requestInviteGroupID).Scan(&reqInviteMemberStatus); err != nil {
		log.Fatalf("verify request_and_invite member status: %v", err)
	}
	if reqInviteMemberStatus != "requested" {
		log.Fatalf("request_and_invite expected requested member, got %s", reqInviteMemberStatus)
	}

	var reqInviteJoinRequestCount int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM group_join_requests WHERE group_id = ? AND user_id = 2 AND status = 'request'`, requestInviteGroupID).Scan(&reqInviteJoinRequestCount); err != nil {
		log.Fatalf("verify request_and_invite join request row: %v", err)
	}
	if reqInviteJoinRequestCount != 1 {
		log.Fatalf("request_and_invite expected one join request row, got %d", reqInviteJoinRequestCount)
	}

	fmt.Println("group join request integration check passed")
}
