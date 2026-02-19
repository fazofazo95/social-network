package main

import (
	"bytes"
	"database/sql"
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
		"description": "permissions check group",
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

func requestJoin(client *http.Client, baseURL string, requesterCookie *http.Cookie, groupID int) error {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/join", baseURL, groupID), nil)
	req.AddCookie(requesterCookie)
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

func assertPendingUnchanged(groupID, requesterID int) error {
	var memberStatus string
	if err := database.DB.QueryRow(`
		SELECT status
		FROM group_members
		WHERE group_id = ? AND user_id = ?
	`, groupID, requesterID).Scan(&memberStatus); err != nil {
		return fmt.Errorf("read group_members: %w", err)
	}
	if memberStatus != "requested" {
		return fmt.Errorf("expected group_members.status=requested got %s", memberStatus)
	}

	var requestStatus, requestType string
	var actionByID sql.NullInt64
	if err := database.DB.QueryRow(`
		SELECT status, request_type, action_by_id
		FROM group_join_requests
		WHERE group_id = ? AND user_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, groupID, requesterID).Scan(&requestStatus, &requestType, &actionByID); err != nil {
		return fmt.Errorf("read group_join_requests: %w", err)
	}
	if requestStatus != "request" {
		return fmt.Errorf("expected group_join_requests.status=request got %s", requestStatus)
	}
	if requestType != "request" {
		return fmt.Errorf("expected group_join_requests.request_type=request got %s", requestType)
	}
	if actionByID.Valid {
		return fmt.Errorf("expected action_by_id to be null, got %d", actionByID.Int64)
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
	requesterCookie, err := login(client, ts.URL, "bob@example.com", "Password123!")
	if err != nil {
		log.Fatalf("requester login: %v", err)
	}
	nonModCookie, err := login(client, ts.URL, "dave@example.com", "Password123!")
	if err != nil {
		log.Fatalf("non-mod login: %v", err)
	}

	acceptGroupID, err := createGroup(client, ts.URL, ownerCookie, fmt.Sprintf("Permissions Accept %d", time.Now().UnixNano()))
	if err != nil {
		log.Fatalf("create accept group: %v", err)
	}
	if err := requestJoin(client, ts.URL, requesterCookie, acceptGroupID); err != nil {
		log.Fatalf("accept scenario request join: %v", err)
	}

	acceptReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/requests/%d/accept", ts.URL, acceptGroupID, 2), nil)
	acceptReq.AddCookie(nonModCookie)
	acceptResp, err := client.Do(acceptReq)
	if err != nil {
		log.Fatalf("non-mod accept call error: %v", err)
	}
	acceptOut, _ := io.ReadAll(acceptResp.Body)
	acceptResp.Body.Close()
	fmt.Printf("non-mod accept -> %d %s\n", acceptResp.StatusCode, string(acceptOut))
	if acceptResp.StatusCode != http.StatusForbidden {
		log.Fatalf("expected non-mod accept to return 403, got %d", acceptResp.StatusCode)
	}
	if err := assertPendingUnchanged(acceptGroupID, 2); err != nil {
		log.Fatalf("accept scenario mutation detected: %v", err)
	}

	rejectGroupID, err := createGroup(client, ts.URL, ownerCookie, fmt.Sprintf("Permissions Reject %d", time.Now().UnixNano()))
	if err != nil {
		log.Fatalf("create reject group: %v", err)
	}
	if err := requestJoin(client, ts.URL, requesterCookie, rejectGroupID); err != nil {
		log.Fatalf("reject scenario request join: %v", err)
	}

	rejectReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/requests/%d/reject", ts.URL, rejectGroupID, 2), nil)
	rejectReq.AddCookie(nonModCookie)
	rejectResp, err := client.Do(rejectReq)
	if err != nil {
		log.Fatalf("non-mod reject call error: %v", err)
	}
	rejectOut, _ := io.ReadAll(rejectResp.Body)
	rejectResp.Body.Close()
	fmt.Printf("non-mod reject -> %d %s\n", rejectResp.StatusCode, string(rejectOut))
	if rejectResp.StatusCode != http.StatusForbidden {
		log.Fatalf("expected non-mod reject to return 403, got %d", rejectResp.StatusCode)
	}
	if err := assertPendingUnchanged(rejectGroupID, 2); err != nil {
		log.Fatalf("reject scenario mutation detected: %v", err)
	}

	fmt.Println("group request permissions integration check passed")
}
