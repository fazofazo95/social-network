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

type listMember struct {
	ID             int    `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	ProfilePicture string `json:"profile_picture"`
	Role           string `json:"group_status"`
}

type listPending struct {
	ID             int    `json:"id"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	ProfilePicture string `json:"profile_picture"`
	Type           string `json:"type"`
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
		"name":        fmt.Sprintf("Group Lists %d", time.Now().UnixNano()),
		"description": "group lists integration",
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
		log.Fatalf("send invite call: %v", err)
	}
	inviteOut, _ := io.ReadAll(inviteResp.Body)
	inviteResp.Body.Close()
	if inviteResp.StatusCode != http.StatusOK {
		log.Fatalf("invite failed: %d %s", inviteResp.StatusCode, string(inviteOut))
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

	membersReq, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/groups/%d/members", ts.URL, groupID), nil)
	membersReq.AddCookie(daveCookie)
	membersResp, err := client.Do(membersReq)
	if err != nil {
		log.Fatalf("members list call: %v", err)
	}
	membersOut, _ := io.ReadAll(membersResp.Body)
	membersResp.Body.Close()
	fmt.Printf("members list -> %d %s\n", membersResp.StatusCode, string(membersOut))
	if membersResp.StatusCode != http.StatusOK {
		log.Fatalf("expected members list 200 got %d", membersResp.StatusCode)
	}
	var membersEnv apiEnvelope
	if err := json.Unmarshal(membersOut, &membersEnv); err != nil {
		log.Fatalf("decode members response: %v", err)
	}
	var members []listMember
	if err := json.Unmarshal(membersEnv.Data, &members); err != nil {
		log.Fatalf("decode members data: %v", err)
	}
	if len(members) != 2 {
		log.Fatalf("expected 2 active members (owner+moderator), got %d", len(members))
	}

	pendingReqsReq, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/groups/%d/requests/pending", ts.URL, groupID), nil)
	pendingReqsReq.AddCookie(aliceCookie)
	pendingReqsResp, err := client.Do(pendingReqsReq)
	if err != nil {
		log.Fatalf("pending requests list call: %v", err)
	}
	pendingReqsOut, _ := io.ReadAll(pendingReqsResp.Body)
	pendingReqsResp.Body.Close()
	fmt.Printf("pending requests list (owner) -> %d %s\n", pendingReqsResp.StatusCode, string(pendingReqsOut))
	if pendingReqsResp.StatusCode != http.StatusOK {
		log.Fatalf("expected pending requests list 200 got %d", pendingReqsResp.StatusCode)
	}
	var pendingReqsEnv apiEnvelope
	if err := json.Unmarshal(pendingReqsOut, &pendingReqsEnv); err != nil {
		log.Fatalf("decode pending requests response: %v", err)
	}
	var pendingReqs []listPending
	if err := json.Unmarshal(pendingReqsEnv.Data, &pendingReqs); err != nil {
		log.Fatalf("decode pending requests data: %v", err)
	}
	if len(pendingReqs) != 1 || pendingReqs[0].Type != "requested" {
		log.Fatalf("expected one pending requested item, got %+v", pendingReqs)
	}

	pendingInvitesReq, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/groups/%d/invites/pending", ts.URL, groupID), nil)
	pendingInvitesReq.AddCookie(bobCookie)
	pendingInvitesResp, err := client.Do(pendingInvitesReq)
	if err != nil {
		log.Fatalf("pending invites list call: %v", err)
	}
	pendingInvitesOut, _ := io.ReadAll(pendingInvitesResp.Body)
	pendingInvitesResp.Body.Close()
	fmt.Printf("pending invites list (moderator) -> %d %s\n", pendingInvitesResp.StatusCode, string(pendingInvitesOut))
	if pendingInvitesResp.StatusCode != http.StatusOK {
		log.Fatalf("expected pending invites list 200 got %d", pendingInvitesResp.StatusCode)
	}
	var pendingInvitesEnv apiEnvelope
	if err := json.Unmarshal(pendingInvitesOut, &pendingInvitesEnv); err != nil {
		log.Fatalf("decode pending invites response: %v", err)
	}
	var pendingInvites []listPending
	if err := json.Unmarshal(pendingInvitesEnv.Data, &pendingInvites); err != nil {
		log.Fatalf("decode pending invites data: %v", err)
	}
	if len(pendingInvites) != 1 || pendingInvites[0].Type != "invited" {
		log.Fatalf("expected one pending invited item, got %+v", pendingInvites)
	}

	nonModPendingReqsReq, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/groups/%d/requests/pending", ts.URL, groupID), nil)
	nonModPendingReqsReq.AddCookie(daveCookie)
	nonModPendingReqsResp, err := client.Do(nonModPendingReqsReq)
	if err != nil {
		log.Fatalf("non-mod pending requests list call: %v", err)
	}
	nonModPendingReqsOut, _ := io.ReadAll(nonModPendingReqsResp.Body)
	nonModPendingReqsResp.Body.Close()
	fmt.Printf("pending requests list (non-mod) -> %d %s\n", nonModPendingReqsResp.StatusCode, string(nonModPendingReqsOut))
	if nonModPendingReqsResp.StatusCode != http.StatusForbidden {
		log.Fatalf("expected non-mod pending requests list 403 got %d", nonModPendingReqsResp.StatusCode)
	}

	nonModPendingInvitesReq, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/groups/%d/invites/pending", ts.URL, groupID), nil)
	nonModPendingInvitesReq.AddCookie(daveCookie)
	nonModPendingInvitesResp, err := client.Do(nonModPendingInvitesReq)
	if err != nil {
		log.Fatalf("non-mod pending invites list call: %v", err)
	}
	nonModPendingInvitesOut, _ := io.ReadAll(nonModPendingInvitesResp.Body)
	nonModPendingInvitesResp.Body.Close()
	fmt.Printf("pending invites list (non-mod) -> %d %s\n", nonModPendingInvitesResp.StatusCode, string(nonModPendingInvitesOut))
	if nonModPendingInvitesResp.StatusCode != http.StatusForbidden {
		log.Fatalf("expected non-mod pending invites list 403 got %d", nonModPendingInvitesResp.StatusCode)
	}

	fmt.Println("group lists integration check passed")
}
