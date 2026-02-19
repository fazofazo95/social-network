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

func loginAndGet(client *http.Client, baseURL, email, password string) (int, string, error) {
	loginPayload := map[string]string{"email": email, "password": password}
	b, _ := json.Marshal(loginPayload)

	loginReq, _ := http.NewRequest(http.MethodPost, baseURL+"/api/login", bytes.NewReader(b))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := client.Do(loginReq)
	if err != nil {
		return 0, "", err
	}
	defer loginResp.Body.Close()

	if loginResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(loginResp.Body)
		return loginResp.StatusCode, string(body), fmt.Errorf("login failed")
	}

	var sessionCookie *http.Cookie
	for _, cookie := range loginResp.Cookies() {
		if cookie.Name == "session_id" {
			sessionCookie = cookie
			break
		}
	}
	if sessionCookie == nil {
		return 0, "", fmt.Errorf("missing session_id cookie after login")
	}

	getReq, _ := http.NewRequest(http.MethodGet, baseURL+"/api/users/settings/content", nil)
	getReq.AddCookie(sessionCookie)
	getResp, err := client.Do(getReq)
	if err != nil {
		return 0, "", err
	}
	defer getResp.Body.Close()

	out, _ := io.ReadAll(getResp.Body)
	return getResp.StatusCode, string(out), nil
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

	// user 1: alice
	client1 := ts.Client()
	status1, body1, err := loginAndGet(client1, ts.URL, "alice@example.com", "Password123!")
	if err != nil {
		fmt.Printf("user1 error: %v\n", err)
	}
	fmt.Printf("user1 GET /api/users/settings/content -> %d\n%s\n\n", status1, body1)

	// user 5: eve
	client5 := ts.Client()
	status5, body5, err := loginAndGet(client5, ts.URL, "eve@example.com", "Password123!")
	if err != nil {
		fmt.Printf("user5 error: %v\n", err)
	}
	fmt.Printf("user5 GET /api/users/settings/content -> %d\n%s\n", status5, body5)
}
