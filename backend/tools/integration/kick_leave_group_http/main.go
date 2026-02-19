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

func createGroup(client *http.Client, baseURL string, ownerCookie *http.Cookie, name string) (int, error) {
	payload := map[string]string{
		"name":        name,
		"description": "kick/leave integration group",
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

func addActiveMember(groupID, userID int, role string) error {
	if _, err := database.DB.Exec(`
		INSERT INTO group_members (group_id, user_id, role, status)
		VALUES (?, ?, ?, 'active')
	`, groupID, userID, role); err != nil {
		return err
	}
	if _, err := database.DB.Exec(`
		UPDATE groups SET group_members = group_members + 1 WHERE id = ?
	`, groupID); err != nil {
		return err
	}
	return nil
}

func groupMembersCount(groupID int) int {
	var c int
	if err := database.DB.QueryRow(`SELECT group_members FROM groups WHERE id = ?`, groupID).Scan(&c); err != nil {
		log.Fatalf("count group_members for group %d: %v", groupID, err)
	}
	return c
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
	daveCookie, err := login(client, ts.URL, "dave@example.com", "Password123!")
	if err != nil {
		log.Fatalf("dave login: %v", err)
	}

	kickGroupID, err := createGroup(client, ts.URL, aliceCookie, fmt.Sprintf("Kick Flow %d", time.Now().UnixNano()))
	if err != nil {
		log.Fatalf("create kick group: %v", err)
	}
	if err := addActiveMember(kickGroupID, 2, "moderator"); err != nil {
		log.Fatalf("seed moderator for kick: %v", err)
	}
	if err := addActiveMember(kickGroupID, 3, "member"); err != nil {
		log.Fatalf("seed member for kick: %v", err)
	}
	beforeKickCount := groupMembersCount(kickGroupID)

	kickReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/members/%d/kick", ts.URL, kickGroupID, 3), nil)
	kickReq.AddCookie(bobCookie)
	kickResp, err := client.Do(kickReq)
	if err != nil {
		log.Fatalf("kick member request: %v", err)
	}
	kickOut, _ := io.ReadAll(kickResp.Body)
	kickResp.Body.Close()
	fmt.Printf("moderator kicks member -> %d %s\n", kickResp.StatusCode, string(kickOut))
	if kickResp.StatusCode != http.StatusOK {
		log.Fatalf("expected kick status 200 got %d", kickResp.StatusCode)
	}
	var kickedRows int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM group_members WHERE group_id = ? AND user_id = ?`, kickGroupID, 3).Scan(&kickedRows); err != nil {
		log.Fatalf("check kicked member row: %v", err)
	}
	afterKickCount := groupMembersCount(kickGroupID)
	if kickedRows != 0 || afterKickCount != beforeKickCount-1 {
		log.Fatalf("kick result invalid: memberRows=%d beforeCount=%d afterCount=%d", kickedRows, beforeKickCount, afterKickCount)
	}

	kickOwnerReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/members/%d/kick", ts.URL, kickGroupID, 1), nil)
	kickOwnerReq.AddCookie(bobCookie)
	kickOwnerResp, err := client.Do(kickOwnerReq)
	if err != nil {
		log.Fatalf("kick owner request: %v", err)
	}
	kickOwnerOut, _ := io.ReadAll(kickOwnerResp.Body)
	kickOwnerResp.Body.Close()
	fmt.Printf("moderator kicks owner -> %d %s\n", kickOwnerResp.StatusCode, string(kickOwnerOut))
	if kickOwnerResp.StatusCode != http.StatusForbidden {
		log.Fatalf("expected kick owner status 403 got %d", kickOwnerResp.StatusCode)
	}

	nonOwnerLeaveGroupID, err := createGroup(client, ts.URL, aliceCookie, fmt.Sprintf("Member Leave %d", time.Now().UnixNano()))
	if err != nil {
		log.Fatalf("create non-owner leave group: %v", err)
	}
	if err := addActiveMember(nonOwnerLeaveGroupID, 2, "member"); err != nil {
		log.Fatalf("seed member for leave: %v", err)
	}
	leaveReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/leave", ts.URL, nonOwnerLeaveGroupID), nil)
	leaveReq.AddCookie(bobCookie)
	leaveResp, err := client.Do(leaveReq)
	if err != nil {
		log.Fatalf("member leave request: %v", err)
	}
	leaveOut, _ := io.ReadAll(leaveResp.Body)
	leaveResp.Body.Close()
	fmt.Printf("member leaves group -> %d %s\n", leaveResp.StatusCode, string(leaveOut))
	if leaveResp.StatusCode != http.StatusOK {
		log.Fatalf("expected member leave status 200 got %d", leaveResp.StatusCode)
	}
	if groupMembersCount(nonOwnerLeaveGroupID) != 1 {
		log.Fatalf("expected group_members=1 after member leave, got %d", groupMembersCount(nonOwnerLeaveGroupID))
	}

	ownerLeavesModGroupID, err := createGroup(client, ts.URL, aliceCookie, fmt.Sprintf("Owner Leave Mod %d", time.Now().UnixNano()))
	if err != nil {
		log.Fatalf("create owner-leaves-mod group: %v", err)
	}
	if err := addActiveMember(ownerLeavesModGroupID, 2, "moderator"); err != nil {
		log.Fatalf("seed moderator for owner leave: %v", err)
	}
	ownerLeaveReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/leave", ts.URL, ownerLeavesModGroupID), nil)
	ownerLeaveReq.AddCookie(aliceCookie)
	ownerLeaveResp, err := client.Do(ownerLeaveReq)
	if err != nil {
		log.Fatalf("owner leave with moderator request: %v", err)
	}
	ownerLeaveOut, _ := io.ReadAll(ownerLeaveResp.Body)
	ownerLeaveResp.Body.Close()
	fmt.Printf("owner leaves (moderator exists) -> %d %s\n", ownerLeaveResp.StatusCode, string(ownerLeaveOut))
	if ownerLeaveResp.StatusCode != http.StatusOK {
		log.Fatalf("expected owner leave status 200 got %d", ownerLeaveResp.StatusCode)
	}
	var ownerID int
	if err := database.DB.QueryRow(`SELECT owner_id FROM groups WHERE id = ?`, ownerLeavesModGroupID).Scan(&ownerID); err != nil {
		log.Fatalf("read new owner after moderator transfer: %v", err)
	}
	if ownerID != 2 {
		log.Fatalf("expected new owner 2 after moderator transfer, got %d", ownerID)
	}

	ownerLeavesMemberGroupID, err := createGroup(client, ts.URL, aliceCookie, fmt.Sprintf("Owner Leave Member %d", time.Now().UnixNano()))
	if err != nil {
		log.Fatalf("create owner-leaves-member group: %v", err)
	}
	if err := addActiveMember(ownerLeavesMemberGroupID, 3, "member"); err != nil {
		log.Fatalf("seed member for owner leave: %v", err)
	}
	ownerLeaveReq2, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/leave", ts.URL, ownerLeavesMemberGroupID), nil)
	ownerLeaveReq2.AddCookie(aliceCookie)
	ownerLeaveResp2, err := client.Do(ownerLeaveReq2)
	if err != nil {
		log.Fatalf("owner leave with member request: %v", err)
	}
	ownerLeaveOut2, _ := io.ReadAll(ownerLeaveResp2.Body)
	ownerLeaveResp2.Body.Close()
	fmt.Printf("owner leaves (no moderator, member exists) -> %d %s\n", ownerLeaveResp2.StatusCode, string(ownerLeaveOut2))
	if ownerLeaveResp2.StatusCode != http.StatusOK {
		log.Fatalf("expected owner leave2 status 200 got %d", ownerLeaveResp2.StatusCode)
	}
	if err := database.DB.QueryRow(`SELECT owner_id FROM groups WHERE id = ?`, ownerLeavesMemberGroupID).Scan(&ownerID); err != nil {
		log.Fatalf("read new owner after member transfer: %v", err)
	}
	if ownerID != 3 {
		log.Fatalf("expected new owner 3 after member transfer, got %d", ownerID)
	}

	ownerSoloGroupID, err := createGroup(client, ts.URL, aliceCookie, fmt.Sprintf("Owner Leave Solo %d", time.Now().UnixNano()))
	if err != nil {
		log.Fatalf("create owner solo group: %v", err)
	}
	ownerLeaveReq3, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/leave", ts.URL, ownerSoloGroupID), nil)
	ownerLeaveReq3.AddCookie(aliceCookie)
	ownerLeaveResp3, err := client.Do(ownerLeaveReq3)
	if err != nil {
		log.Fatalf("owner leave solo request: %v", err)
	}
	ownerLeaveOut3, _ := io.ReadAll(ownerLeaveResp3.Body)
	ownerLeaveResp3.Body.Close()
	fmt.Printf("owner leaves (solo) -> %d %s\n", ownerLeaveResp3.StatusCode, string(ownerLeaveOut3))
	if ownerLeaveResp3.StatusCode != http.StatusOK {
		log.Fatalf("expected owner leave3 status 200 got %d", ownerLeaveResp3.StatusCode)
	}
	var groupRows int
	if err := database.DB.QueryRow(`SELECT COUNT(*) FROM groups WHERE id = ?`, ownerSoloGroupID).Scan(&groupRows); err != nil {
		log.Fatalf("check solo group delete: %v", err)
	}
	if groupRows != 0 {
		log.Fatalf("expected solo group to be deleted, rows=%d", groupRows)
	}

	nonMemberLeaveGroupID, err := createGroup(client, ts.URL, aliceCookie, fmt.Sprintf("Non Member Leave %d", time.Now().UnixNano()))
	if err != nil {
		log.Fatalf("create non-member leave group: %v", err)
	}
	nonMemberLeaveReq, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/groups/%d/leave", ts.URL, nonMemberLeaveGroupID), nil)
	nonMemberLeaveReq.AddCookie(daveCookie)
	nonMemberLeaveResp, err := client.Do(nonMemberLeaveReq)
	if err != nil {
		log.Fatalf("non-member leave request: %v", err)
	}
	nonMemberLeaveOut, _ := io.ReadAll(nonMemberLeaveResp.Body)
	nonMemberLeaveResp.Body.Close()
	fmt.Printf("non-member leaves group -> %d %s\n", nonMemberLeaveResp.StatusCode, string(nonMemberLeaveOut))
	if nonMemberLeaveResp.StatusCode != http.StatusNotFound {
		log.Fatalf("expected non-member leave status 404 got %d", nonMemberLeaveResp.StatusCode)
	}

	fmt.Println("kick/leave group integration check passed")
}
