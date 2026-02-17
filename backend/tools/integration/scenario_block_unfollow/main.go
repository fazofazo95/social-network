package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"time"

	database "backend/pkg/db/sqlite"
	handlers "backend/pkg/handlers"
	"backend/pkg/middleware"
)

func printResp(resp *http.Response) {
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("HTTP/1.1 %d %s\n", resp.StatusCode, http.StatusText(resp.StatusCode))
	for k, v := range resp.Header {
		fmt.Printf("%s: %s\n", k, v)
	}
	fmt.Printf("\n%s\n\n", string(body))
}

func main() {
	// init DB
	dbPath := "pkg/db/social_network.db"
	if err := database.Init(dbPath); err != nil {
		log.Fatalf("init db: %v", err)
	}
	defer database.DB.Close()

	mux := http.NewServeMux()
	handlers.UserRoutes(mux)
	handlers.AuthRoutes(mux)
	handlers.PostRoutes(mux, database.DB)
	handlers.CommentRoutes(mux, database.DB)
	handlers.FeedRoutes(mux)

	handlerWithCORS := middleware.CorsMiddleware(mux)
	ts := httptest.NewServer(handlerWithCORS)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 30 * time.Second}

	// login as alice (user 1)
	loginURL := ts.URL + "/api/login"
	loginBody := map[string]string{"email": "alice@example.com", "password": "Password123!"}
	b, _ := json.Marshal(loginBody)
	resp, err := client.Post(loginURL, "application/json", bytes.NewReader(b))
	if err != nil {
		log.Fatalf("login request failed: %v", err)
	}
	fmt.Println("--- LOGIN RESPONSE ---")
	printResp(resp)

	// Block users 2,3,4
	for _, id := range []int{2, 3, 4} {
		url := fmt.Sprintf("%s/api/users/%d/block", ts.URL, id)
		req, _ := http.NewRequest("POST", url, nil)
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("block %d failed: %v", id, err)
			continue
		}
		fmt.Printf("--- BLOCK %d RESPONSE ---\n", id)
		printResp(resp)
	}

	// Unfollow users 5 and 6
	for _, id := range []int{5, 6} {
		url := fmt.Sprintf("%s/api/users/%d/unfollow", ts.URL, id)
		req, _ := http.NewRequest("DELETE", url, nil)
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("unfollow %d failed: %v", id, err)
			continue
		}
		fmt.Printf("--- UNFOLLOW %d RESPONSE ---\n", id)
		printResp(resp)
	}

	// Discover
	discoverURL := ts.URL + "/api/discover"
	resp, err = client.Get(discoverURL)
	if err != nil {
		log.Fatalf("discover request failed: %v", err)
	}
	fmt.Println("--- DISCOVER RESPONSE ---")
	printResp(resp)

	_ = time.Now()
}
