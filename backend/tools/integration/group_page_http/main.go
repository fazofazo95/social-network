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
		"name":        fmt.Sprintf("Group Page %d", time.Now().UnixNano()),
		"description": "group page integration",
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

	inviteReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/invite/%d", ts.URL, groupID, 2), nil)
	inviteReq.AddCookie(aliceCookie)
	inviteResp, err := client.Do(inviteReq)
	if err != nil {
		log.Fatalf("send invite to bob: %v", err)
	}
	inviteOut, _ := io.ReadAll(inviteResp.Body)
	inviteResp.Body.Close()
	if inviteResp.StatusCode != http.StatusOK {
		log.Fatalf("invite failed: %d %s", inviteResp.StatusCode, string(inviteOut))
	}

	requestReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/join", ts.URL, groupID), nil)
	requestReq.AddCookie(carolCookie)
	requestResp, err := client.Do(requestReq)
	if err != nil {
		log.Fatalf("send request from carol: %v", err)
	}
	requestOut, _ := io.ReadAll(requestResp.Body)
	requestResp.Body.Close()
	if requestResp.StatusCode != http.StatusOK {
		log.Fatalf("request failed: %d %s", requestResp.StatusCode, string(requestOut))
	}

	for _, tc := range []struct {
		label  string
		cookie *http.Cookie
	}{
		{label: "public/no-pending(can request)", cookie: daveCookie},
		{label: "public/pending-invite", cookie: bobCookie},
		{label: "public/pending-request", cookie: carolCookie},
	} {
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/groups/%d", ts.URL, groupID), nil)
		req.AddCookie(tc.cookie)
		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("group page (%s) call failed: %v", tc.label, err)
		}
		out, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("%s -> %d\n%s\n\n", tc.label, resp.StatusCode, string(out))
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("expected 200 for %s got %d", tc.label, resp.StatusCode)
		}
	}

	fmt.Println("group page integration check passed")
}
