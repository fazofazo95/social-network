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
		"name":        fmt.Sprintf("Update Settings %d", time.Now().UnixNano()),
		"description": "settings integration group",
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

	groupID, err := createGroup(client, ts.URL, aliceCookie)
	if err != nil {
		log.Fatalf("create group: %v", err)
	}

	ownerPayload := map[string]string{"visibility": "private", "join_mode": "request_and_invite"}
	ownerBody, _ := json.Marshal(ownerPayload)
	ownerReq, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/groups/%d/settings", ts.URL, groupID), bytes.NewReader(ownerBody))
	ownerReq.Header.Set("Content-Type", "application/json")
	ownerReq.AddCookie(aliceCookie)
	ownerResp, err := client.Do(ownerReq)
	if err != nil {
		log.Fatalf("owner update settings call: %v", err)
	}
	ownerOut, _ := io.ReadAll(ownerResp.Body)
	ownerResp.Body.Close()
	fmt.Printf("owner updates settings -> %d %s\n", ownerResp.StatusCode, string(ownerOut))
	if ownerResp.StatusCode != http.StatusOK {
		log.Fatalf("expected owner settings update 200 got %d", ownerResp.StatusCode)
	}

	var visibility, joinMode string
	if err := database.DB.QueryRow(`SELECT visibility, join_mode FROM groups WHERE id = ?`, groupID).Scan(&visibility, &joinMode); err != nil {
		log.Fatalf("verify updated settings from DB: %v", err)
	}
	if visibility != "private" || joinMode != "request_and_invite" {
		log.Fatalf("unexpected settings after owner update: visibility=%s join_mode=%s", visibility, joinMode)
	}

	ownerGetReq, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/groups/%d/settings", ts.URL, groupID), nil)
	ownerGetReq.AddCookie(aliceCookie)
	ownerGetResp, err := client.Do(ownerGetReq)
	if err != nil {
		log.Fatalf("owner get settings call: %v", err)
	}
	ownerGetOut, _ := io.ReadAll(ownerGetResp.Body)
	ownerGetResp.Body.Close()
	fmt.Printf("owner gets settings -> %d %s\n", ownerGetResp.StatusCode, string(ownerGetOut))
	if ownerGetResp.StatusCode != http.StatusOK {
		log.Fatalf("expected owner get settings 200 got %d", ownerGetResp.StatusCode)
	}

	nonOwnerPayload := map[string]string{"visibility": "public"}
	nonOwnerBody, _ := json.Marshal(nonOwnerPayload)
	nonOwnerReq, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/groups/%d/settings", ts.URL, groupID), bytes.NewReader(nonOwnerBody))
	nonOwnerReq.Header.Set("Content-Type", "application/json")
	nonOwnerReq.AddCookie(bobCookie)
	nonOwnerResp, err := client.Do(nonOwnerReq)
	if err != nil {
		log.Fatalf("non-owner update settings call: %v", err)
	}
	nonOwnerOut, _ := io.ReadAll(nonOwnerResp.Body)
	nonOwnerResp.Body.Close()
	fmt.Printf("non-owner updates settings -> %d %s\n", nonOwnerResp.StatusCode, string(nonOwnerOut))
	if nonOwnerResp.StatusCode != http.StatusForbidden {
		log.Fatalf("expected non-owner settings update 403 got %d", nonOwnerResp.StatusCode)
	}

	nonOwnerGetReq, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/groups/%d/settings", ts.URL, groupID), nil)
	nonOwnerGetReq.AddCookie(bobCookie)
	nonOwnerGetResp, err := client.Do(nonOwnerGetReq)
	if err != nil {
		log.Fatalf("non-owner get settings call: %v", err)
	}
	nonOwnerGetOut, _ := io.ReadAll(nonOwnerGetResp.Body)
	nonOwnerGetResp.Body.Close()
	fmt.Printf("non-owner gets settings -> %d %s\n", nonOwnerGetResp.StatusCode, string(nonOwnerGetOut))
	if nonOwnerGetResp.StatusCode != http.StatusForbidden {
		log.Fatalf("expected non-owner get settings 403 got %d", nonOwnerGetResp.StatusCode)
	}

	invalidPayload := map[string]string{"join_mode": "whatever_mode"}
	invalidBody, _ := json.Marshal(invalidPayload)
	invalidReq, _ := http.NewRequest(http.MethodPatch, fmt.Sprintf("%s/api/groups/%d/settings", ts.URL, groupID), bytes.NewReader(invalidBody))
	invalidReq.Header.Set("Content-Type", "application/json")
	invalidReq.AddCookie(aliceCookie)
	invalidResp, err := client.Do(invalidReq)
	if err != nil {
		log.Fatalf("invalid settings call: %v", err)
	}
	invalidOut, _ := io.ReadAll(invalidResp.Body)
	invalidResp.Body.Close()
	fmt.Printf("owner sends invalid join_mode -> %d %s\n", invalidResp.StatusCode, string(invalidOut))
	if invalidResp.StatusCode != http.StatusBadRequest {
		log.Fatalf("expected invalid settings update 400 got %d", invalidResp.StatusCode)
	}

	fmt.Println("update group settings integration check passed")
}
