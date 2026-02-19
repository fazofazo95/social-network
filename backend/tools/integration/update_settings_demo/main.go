package main

import (
	"bytes"
	"context"
	"database/sql"
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
	client := &http.Client{Jar: jar, Timeout: 10 * time.Second}

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

	// 2) PATCH settings: set phone_number_vis to visible
	patchURL := ts.URL + "/api/users/settings"
	payload := map[string]interface{}{"phone_number_vis": "visible"}
	pb, _ := json.Marshal(payload)
	req, _ := http.NewRequest("PATCH", patchURL, bytes.NewReader(pb))
	req.Header.Set("Content-Type", "application/json")
	resp, err = client.Do(req)
	if err != nil {
		log.Fatalf("patch request failed: %v", err)
	}
	fmt.Println("--- PATCH RESPONSE ---")
	printResp(resp)

	// 3) GET settings
	getURL := ts.URL + "/api/users/settings"
	resp, err = client.Get(getURL)
	if err != nil {
		log.Fatalf("get settings failed: %v", err)
	}
	fmt.Println("--- GET SETTINGS RESPONSE ---")
	printResp(resp)

	// 4) Show raw DB row for user id 1
	var phoneVis sql.NullInt64
	row := database.DB.QueryRowContext(ctx, "SELECT phone_number_vis FROM user_settings WHERE id = ?", 1)
	if err := row.Scan(&phoneVis); err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("DB: no user_settings row for user 1")
		} else {
			fmt.Printf("DB query error: %v\n", err)
		}
	} else {
		fmt.Printf("DB: phone_number_vis for user 1 = %v\n", phoneVis.Int64)
	}

	_ = ctx
}
