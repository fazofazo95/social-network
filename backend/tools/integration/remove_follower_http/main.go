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

func ensureAcceptedFollow(followerID, followedID int) error {
	_, err := database.DB.Exec(`
		INSERT INTO followers (follower_id, followed_id, status)
		VALUES (?, ?, 'accepted')
		ON CONFLICT(follower_id, followed_id)
		DO UPDATE SET status = 'accepted', updated_at = CURRENT_TIMESTAMP
	`, followerID, followedID)
	return err
}

func countRelationship(followerID, followedID int) (int, error) {
	var c int
	err := database.DB.QueryRow(`
		SELECT COUNT(*)
		FROM followers
		WHERE follower_id = ? AND followed_id = ? AND status = 'accepted'
	`, followerID, followedID).Scan(&c)
	return c, err
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

	// bob follows alice (2 -> 1) accepted
	if err := ensureAcceptedFollow(2, 1); err != nil {
		log.Fatalf("seed accepted follow: %v", err)
	}

	aliceCookie, err := login(client, ts.URL, "alice@example.com", "Password123!")
	if err != nil {
		log.Fatalf("alice login: %v", err)
	}

	before, err := countRelationship(2, 1)
	if err != nil {
		log.Fatalf("count relationship before: %v", err)
	}
	if before != 1 {
		log.Fatalf("expected relationship count 1 before removal, got %d", before)
	}

	removeReq, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/users/2/remove-follower", nil)
	removeReq.AddCookie(aliceCookie)
	removeResp, err := client.Do(removeReq)
	if err != nil {
		log.Fatalf("remove follower request failed: %v", err)
	}
	removeOut, _ := io.ReadAll(removeResp.Body)
	removeResp.Body.Close()
	fmt.Printf("remove follower -> %d %s\n", removeResp.StatusCode, string(removeOut))
	if removeResp.StatusCode != http.StatusOK {
		log.Fatalf("expected remove follower status 200, got %d", removeResp.StatusCode)
	}

	after, err := countRelationship(2, 1)
	if err != nil {
		log.Fatalf("count relationship after: %v", err)
	}
	if after != 0 {
		log.Fatalf("expected relationship count 0 after removal, got %d", after)
	}

	removeAgainReq, _ := http.NewRequest(http.MethodDelete, ts.URL+"/api/users/2/remove-follower", nil)
	removeAgainReq.AddCookie(aliceCookie)
	removeAgainResp, err := client.Do(removeAgainReq)
	if err != nil {
		log.Fatalf("remove follower again request failed: %v", err)
	}
	removeAgainOut, _ := io.ReadAll(removeAgainResp.Body)
	removeAgainResp.Body.Close()
	fmt.Printf("remove follower again -> %d %s\n", removeAgainResp.StatusCode, string(removeAgainOut))
	if removeAgainResp.StatusCode != http.StatusNotFound {
		log.Fatalf("expected remove follower again status 404, got %d", removeAgainResp.StatusCode)
	}

	fmt.Println("remove follower integration check passed")
}
