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

func getProfile(client *http.Client, baseURL string, cookie *http.Cookie, userID int) (int, string, error) {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/users/%d", baseURL, userID), nil)
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

	for id := 2; id <= 6; id++ {
		status, body, err := getProfile(client, ts.URL, cookie, id)
		if err != nil {
			fmt.Printf("GET /api/users/%d error: %v\n", id, err)
			continue
		}
		fmt.Printf("GET /api/users/%d -> %d\n%s\n\n", id, status, body)
	}
}
