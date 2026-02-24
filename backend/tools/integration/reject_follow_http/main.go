package main

import (
	"bytes"
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
	body, _ := json.Marshal(payload)
	req, _ := http.NewRequest(http.MethodPost, baseURL+"/api/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		out, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("login failed (%s): %d %s", email, resp.StatusCode, string(out))
	}

	for _, c := range resp.Cookies() {
		if c.Name == "session_id" {
			return c, nil
		}
	}
	return nil, fmt.Errorf("missing session cookie for %s", email)
}

func follow(client *http.Client, baseURL string, cookie *http.Cookie, targetID int) (int, string) {
	req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/users/%d/follow", baseURL, targetID), nil)
	req.AddCookie(cookie)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("follow request failed: %v", err)
	}
	defer resp.Body.Close()
	out, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, string(out)
}

func countPending(followerID, followedID int) int {
	var c int
	if err := database.DB.QueryRow(`
		SELECT COUNT(*)
		FROM followers
		WHERE follower_id = ? AND followed_id = ? AND status = 'pending'
	`, followerID, followedID).Scan(&c); err != nil {
		log.Fatalf("count pending failed: %v", err)
	}
	return c
}

func countActiveFollowOrPending(followerID, followedID int) int {
	var c int
	if err := database.DB.QueryRow(`
		SELECT COUNT(*)
		FROM followers
		WHERE follower_id = ? AND followed_id = ? AND status IN ('pending', 'accepted')
	`, followerID, followedID).Scan(&c); err != nil {
		log.Fatalf("count active follow/pending failed: %v", err)
	}
	return c
}

func seedPendingFollow(followerID, followedID int) {
	if _, err := database.DB.Exec(`
		INSERT INTO followers (follower_id, followed_id, status)
		VALUES (?, ?, 'pending')
		ON CONFLICT(follower_id, followed_id)
		DO UPDATE SET status = 'pending', updated_at = CURRENT_TIMESTAMP
	`, followerID, followedID); err != nil {
		log.Fatalf("seed pending follow failed: %v", err)
	}
}

func main() {
	if err := database.Init("pkg/db/social_network.db"); err != nil {
		log.Fatalf("init db: %v", err)
	}
	defer database.DB.Close()

	if _, err := database.DB.Exec(`UPDATE users SET profile_type = 1 WHERE id = 1`); err != nil {
		log.Fatalf("normalize alice profile_type: %v", err)
	}

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

	seedPendingFollow(2, 1)
	if countPending(2, 1) != 1 {
		log.Fatalf("expected 1 pending follow request before reject")
	}

	rejectReq, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/users/2/follow/reject", nil)
	rejectReq.AddCookie(aliceCookie)
	rejectResp, err := client.Do(rejectReq)
	if err != nil {
		log.Fatalf("reject follow request failed: %v", err)
	}
	rejectOut, _ := io.ReadAll(rejectResp.Body)
	rejectResp.Body.Close()
	fmt.Printf("alice rejects bob follow -> %d %s\n", rejectResp.StatusCode, string(rejectOut))
	if rejectResp.StatusCode != http.StatusOK {
		log.Fatalf("expected reject follow 200, got %d", rejectResp.StatusCode)
	}
	if countPending(2, 1) != 0 {
		log.Fatalf("expected pending follow request to be removed after reject")
	}

	status2, body2 := follow(client, ts.URL, bobCookie, 1)
	fmt.Printf("bob follow alice again -> %d %s\n", status2, body2)
	if status2 != http.StatusCreated {
		log.Fatalf("expected second follow create 201, got %d", status2)
	}
	if countActiveFollowOrPending(2, 1) != 1 {
		log.Fatalf("expected follow relationship to be recreated after second follow")
	}

	fmt.Println("reject follow integration check passed")
}
