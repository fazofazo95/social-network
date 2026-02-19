package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"

	database "backend/pkg/db/sqlite"
	handlers "backend/pkg/handlers"
)

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

func callGet(client *http.Client, url string, cookie *http.Cookie) (int, string, error) {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.AddCookie(cookie)
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(out), nil
}

func main() {
	if err := database.Init("pkg/db/social_network.db"); err != nil {
		log.Fatalf("init db: %v", err)
	}
	defer database.DB.Close()
	ctx := context.Background()

	// Setup deterministic scenarios for viewer user 1.
	// A) viewer blocked target public (id=2 -> profile_type=0)
	_, _ = database.DB.ExecContext(ctx, "UPDATE users SET profile_type = 0 WHERE id = 2")
	_, _ = database.DB.ExecContext(ctx, "INSERT INTO followers (follower_id, followed_id, status) VALUES (1,2,'blocked') ON CONFLICT(follower_id, followed_id) DO UPDATE SET status='blocked'")
	_, _ = database.DB.ExecContext(ctx, "DELETE FROM followers WHERE follower_id = 2 AND followed_id = 1 AND status='blocked'")

	// B) viewer blocked target private (id=6 -> profile_type=1)
	_, _ = database.DB.ExecContext(ctx, "UPDATE users SET profile_type = 1 WHERE id = 6")
	_, _ = database.DB.ExecContext(ctx, "INSERT INTO followers (follower_id, followed_id, status) VALUES (1,6,'blocked') ON CONFLICT(follower_id, followed_id) DO UPDATE SET status='blocked'")
	_, _ = database.DB.ExecContext(ctx, "DELETE FROM followers WHERE follower_id = 6 AND followed_id = 1 AND status='blocked'")

	// C) target blocked viewer (id=4 blocks user 1)
	_, _ = database.DB.ExecContext(ctx, "INSERT INTO followers (follower_id, followed_id, status) VALUES (4,1,'blocked') ON CONFLICT(follower_id, followed_id) DO UPDATE SET status='blocked'")
	_, _ = database.DB.ExecContext(ctx, "DELETE FROM followers WHERE follower_id = 1 AND followed_id = 4 AND status='blocked'")

	mux := http.NewServeMux()
	handlers.UserRoutes(mux)
	handlers.AuthRoutes(mux)
	handlers.FollowRoutes(mux)
	handlers.PostRoutes(mux, database.DB)
	handlers.CommentRoutes(mux, database.DB)
	handlers.FeedRoutes(mux)

	ts := httptest.NewServer(mux)
	defer ts.Close()

	client := ts.Client()
	cookie, err := login(client, ts.URL, "alice@example.com", "Password123!")
	if err != nil {
		log.Fatalf("login error: %v", err)
	}

	cases := []struct {
		label    string
		targetID int
	}{
		{"viewer_blocked_target_public", 2},
		{"viewer_blocked_target_private", 6},
		{"target_blocked_viewer", 4},
	}

	for _, c := range cases {
		status, body, err := callGet(client, fmt.Sprintf("%s/api/users/%d", ts.URL, c.targetID), cookie)
		if err != nil {
			fmt.Printf("%s error: %v\n", c.label, err)
			continue
		}
		fmt.Printf("%s GET /api/users/%d -> %d\n%s\n\n", c.label, c.targetID, status, body)
	}
}
