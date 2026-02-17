package main

import (
	"bytes"
	"context"
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
	ctx := context.Background()
	// initialize DB (uses existing db file in repo)
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
	// Note: routes are registered by UserRoutes (includes follow endpoints).

	handlerWithCORS := middleware.CorsMiddleware(mux)

	// start in-process server
	ts := httptest.NewServer(handlerWithCORS)
	defer ts.Close()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar, Timeout: 5 * time.Second}

	// 1) login as alice
	loginURL := ts.URL + "/api/login"
	loginBody := map[string]string{"email": "alice@example.com", "password": "Password123!"}
	b, _ := json.Marshal(loginBody)
	resp, err := client.Post(loginURL, "application/json", bytes.NewReader(b))
	if err != nil {
		log.Fatalf("login request failed: %v", err)
	}
	fmt.Println("--- LOGIN RESPONSE ---")
	printResp(resp)

	// 2) discover
	discoverURL := ts.URL + "/api/discover"
	resp, err = client.Get(discoverURL)
	if err != nil {
		log.Fatalf("discover request failed: %v", err)
	}
	fmt.Println("--- DISCOVER RESPONSE (initial) ---")
	printResp(resp)

	// parse discovered ids
	var parsed map[string]interface{}
	resp2, err := client.Get(discoverURL)
	if err != nil {
		log.Fatalf("discover request failed: %v", err)
	}
	body, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	json.Unmarshal(body, &parsed)
	ids := []int{}
	if data, ok := parsed["data"]; ok {
		if arr, ok := data.([]interface{}); ok {
			for _, item := range arr {
				if m, ok := item.(map[string]interface{}); ok {
					if idf, ok := m["id"]; ok {
						if idf != nil {
							// numbers in JSON are float64
							if f, ok := idf.(float64); ok {
								ids = append(ids, int(f))
							}
						}
					}
				}
			}
		}
	}

	fmt.Printf("Discovered ids to follow: %v\n", ids)

	// 3) follow each id
	for _, id := range ids {
		followURL := fmt.Sprintf("%s/api/users/%d/follow", ts.URL, id)
		req, _ := http.NewRequest("POST", followURL, nil)
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("follow %d failed: %v", id, err)
			continue
		}
		fmt.Printf("--- FOLLOW %d RESPONSE ---\n", id)
		printResp(resp)
	}

	// 4) discover again
	resp, err = client.Get(discoverURL)
	if err != nil {
		log.Fatalf("discover request failed: %v", err)
	}
	fmt.Println("--- DISCOVER RESPONSE (after follows) ---")
	printResp(resp)

	// done
	_ = ctx
}
