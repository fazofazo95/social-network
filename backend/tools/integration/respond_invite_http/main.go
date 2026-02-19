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
		"description": "respond invite integration group",
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
	if data.ID <= 0 {
		return 0, fmt.Errorf("invalid group id")
	}
	return data.ID, nil
}

func invite(client *http.Client, baseURL string, inviterCookie *http.Cookie, groupID, targetUserID int) error {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/invite/%d", baseURL, groupID, targetUserID), nil)
	req.AddCookie(inviterCookie)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("invite failed: %d %s", resp.StatusCode, string(out))
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
	ownerCookie, err := login(client, ts.URL, "alice@example.com", "Password123!")
	if err != nil {
		log.Fatalf("owner login: %v", err)
	}
	bobCookie, err := login(client, ts.URL, "bob@example.com", "Password123!")
	if err != nil {
		log.Fatalf("bob login: %v", err)
	}
	carolCookie, err := login(client, ts.URL, "carol@example.com", "Password123!")
	if err != nil {
		log.Fatalf("carol login: %v", err)
	}

	acceptGroupID, err := createGroup(client, ts.URL, ownerCookie, fmt.Sprintf("Accept Invite %d", time.Now().UnixNano()))
	if err != nil {
		log.Fatalf("create accept group: %v", err)
	}
	if err := invite(client, ts.URL, ownerCookie, acceptGroupID, 2); err != nil {
		log.Fatalf("invite bob: %v", err)
	}

	acceptReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/invite/accept", ts.URL, acceptGroupID), nil)
	acceptReq.AddCookie(bobCookie)
	acceptResp, err := client.Do(acceptReq)
	if err != nil {
		log.Fatalf("accept invite call: %v", err)
	}
	acceptOut, _ := io.ReadAll(acceptResp.Body)
	acceptResp.Body.Close()
	fmt.Printf("accept invite -> %d %s\n", acceptResp.StatusCode, string(acceptOut))
	if acceptResp.StatusCode != http.StatusOK {
		log.Fatalf("expected accept invite 200 got %d", acceptResp.StatusCode)
	}

	var status, role, joinedAt string
	if err := database.DB.QueryRow(`
		SELECT status, role, COALESCE(datetime(joined_at), '')
		FROM group_members
		WHERE group_id = ? AND user_id = ?
	`, acceptGroupID, 2).Scan(&status, &role, &joinedAt); err != nil {
		log.Fatalf("verify accepted member row: %v", err)
	}
	if status != "active" || role != "member" || joinedAt == "" {
		log.Fatalf("unexpected accepted member values: status=%s role=%s joined_at=%s", status, role, joinedAt)
	}

	var membersCount int
	if err := database.DB.QueryRow(`SELECT group_members FROM groups WHERE id = ?`, acceptGroupID).Scan(&membersCount); err != nil {
		log.Fatalf("verify accept group_members count: %v", err)
	}
	if membersCount != 2 {
		log.Fatalf("expected group_members=2 after accepting invite, got %d", membersCount)
	}

	var reqStatus, reqType string
	var actionByID int
	if err := database.DB.QueryRow(`
		SELECT status, request_type, action_by_id
		FROM group_join_requests
		WHERE group_id = ? AND user_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, acceptGroupID, 2).Scan(&reqStatus, &reqType, &actionByID); err != nil {
		log.Fatalf("verify accepted invite request row: %v", err)
	}
	if reqStatus != "accepted" || reqType != "invite" || actionByID != 2 {
		log.Fatalf("unexpected accepted invite row: status=%s type=%s action_by_id=%d", reqStatus, reqType, actionByID)
	}

	rejectGroupID, err := createGroup(client, ts.URL, ownerCookie, fmt.Sprintf("Reject Invite %d", time.Now().UnixNano()))
	if err != nil {
		log.Fatalf("create reject group: %v", err)
	}
	if err := invite(client, ts.URL, ownerCookie, rejectGroupID, 3); err != nil {
		log.Fatalf("invite carol: %v", err)
	}

	rejectReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/invite/reject", ts.URL, rejectGroupID), nil)
	rejectReq.AddCookie(carolCookie)
	rejectResp, err := client.Do(rejectReq)
	if err != nil {
		log.Fatalf("reject invite call: %v", err)
	}
	rejectOut, _ := io.ReadAll(rejectResp.Body)
	rejectResp.Body.Close()
	fmt.Printf("reject invite -> %d %s\n", rejectResp.StatusCode, string(rejectOut))
	if rejectResp.StatusCode != http.StatusOK {
		log.Fatalf("expected reject invite 200 got %d", rejectResp.StatusCode)
	}

	var rejectedMemberRows int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM group_members WHERE group_id = ? AND user_id = ?`, rejectGroupID, 3).Scan(&rejectedMemberRows); err != nil {
		log.Fatalf("verify rejected member row count: %v", err)
	}
	if rejectedMemberRows != 0 {
		log.Fatalf("expected no group_members row after reject invite, got %d", rejectedMemberRows)
	}

	if err := database.DB.QueryRow(`SELECT group_members FROM groups WHERE id = ?`, rejectGroupID).Scan(&membersCount); err != nil {
		log.Fatalf("verify reject group_members count: %v", err)
	}
	if membersCount != 1 {
		log.Fatalf("expected group_members=1 after rejecting invite, got %d", membersCount)
	}

	if err := database.DB.QueryRow(`
		SELECT status, request_type, action_by_id
		FROM group_join_requests
		WHERE group_id = ? AND user_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, rejectGroupID, 3).Scan(&reqStatus, &reqType, &actionByID); err != nil {
		log.Fatalf("verify rejected invite request row: %v", err)
	}
	if reqStatus != "rejected" || reqType != "invite" || actionByID != 3 {
		log.Fatalf("unexpected rejected invite row: status=%s type=%s action_by_id=%d", reqStatus, reqType, actionByID)
	}

	fmt.Println("respond invite integration check passed")
}
