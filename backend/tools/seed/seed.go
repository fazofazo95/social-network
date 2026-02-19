package seed

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	queries "backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
	"backend/pkg/models"

	"golang.org/x/crypto/bcrypt"
)

type SeedUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SeedFromJSON reads the given JSON file and creates users via queries.SignUp.
// It returns the number of users created or an error.
func SeedFromJSON(path string) (int, error) {
	fmt.Printf("seed: SeedFromJSON start path=%s\n", path)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}

	var users []SeedUser
	if err := json.Unmarshal(b, &users); err != nil {
		return 0, err
	}

	dbPath := "pkg/db/social_network.db"
	if err := database.Init(dbPath); err != nil {
		return 0, fmt.Errorf("init db: %w", err)
	}
	fmt.Println("seed: SeedFromJSON database initialized")

	// Ensure `level` column exists (some older DBs may lack it); add if missing.
	{
		rows, err := database.DB.QueryContext(context.Background(), `PRAGMA table_info(users);`)
		if err == nil {
			defer rows.Close()
			hasLevel := false
			for rows.Next() {
				var cid int
				var name string
				var ctype string
				var notnull int
				var dfltValue sql.NullString
				var pk int
				if err := rows.Scan(&cid, &name, &ctype, &notnull, &dfltValue, &pk); err == nil {
					if name == "level" {
						hasLevel = true
						break
					}
				}
			}
			if !hasLevel {
				fmt.Println("seed: adding missing 'level' column to users table")
				_, _ = database.DB.ExecContext(context.Background(), `ALTER TABLE users ADD COLUMN level TEXT;`)
			}
		}
	}

	created := 0
	for _, u := range users {
		// Create mock profile fields from username
		first := ""
		last := ""
		if u.Username != "" {
			// split username on non-alpha chars or use as first name
			first = u.Username
			last = "User"
		}
		// default birthday
		birthday := "1990-01-01"
		avatar := ""
		nickname := u.Username
		about := ""

		input := models.Signup_fields{
			Email:     u.Email,
			Username:  u.Username,
			Password:  u.Password,
			FirstName: first,
			LastName:  last,
			Birthday:  birthday,
			Avatar:    avatar,
			Nickname:  nickname,
			AboutMe:   about,
		}
		// Use a lightweight insert for seeding to avoid transactional locking issues.
		// Hash password
		log.Printf("seed: hashing password for %s", u.Email)
		// Use a low cost for seeding to avoid long CPU-bound delays
		hash, herr := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.MinCost)
		if herr != nil {
			log.Printf("seed: bcrypt error for %s: %v", u.Email, herr)
			continue
		}
		log.Printf("seed: bcrypt hash generated for %s (len=%d)", u.Email, len(hash))

		// try insert into login_users
		res, ierr := database.DB.ExecContext(context.Background(), `INSERT INTO login_users (email, username, password_hash) VALUES (?, ?, ?);`, u.Email, u.Username, string(hash))
		if ierr != nil {
			// if duplicate, skip and continue
			log.Printf("seed: login_users insert error for %s: %v", u.Email, ierr)
			continue
		}
		log.Printf("seed: inserted login_users email=%s", u.Email)

		var userID int64
		if id, err := res.LastInsertId(); err == nil && id > 0 {
			userID = id
			log.Printf("seed: got lastinsertid=%d for email=%s", userID, u.Email)
		} else {
			// fallback to SELECT id
			var uid int
			row := database.DB.QueryRowContext(context.Background(), "SELECT id FROM login_users WHERE email = ?;", u.Email)
			if rowErr := row.Scan(&uid); rowErr != nil {
				log.Printf("seed: unable to resolve id for %s: %v", u.Email, rowErr)
				continue
			}
			userID = int64(uid)
		}

		// insert into users (match current migrations schema)
		_, uerr := database.DB.ExecContext(context.Background(), `INSERT INTO users (id, first_name, last_name, birthday_date, relationship_status, employed_at, phone_number, profile_picture, pictures, level) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`, userID, input.FirstName, input.LastName, input.Birthday, nil, nil, nil, input.Avatar, nil, "basic")
		if uerr != nil {
			log.Printf("seed: users insert error for id=%d: %v", userID, uerr)
			continue
		}

		// create default user_settings row (if missing)
		if serr := queries.InsertDefaultUserSettings(context.Background(), database.DB, int(userID)); serr != nil {
			log.Printf("seed: failed to create default user_settings for id=%d: %v", userID, serr)
		}
		log.Printf("seed: inserted users id=%d", userID)
		created++
		fmt.Printf("seed: SeedFromJSON created user username=%s email=%s\n", u.Username, u.Email)
	}

	fmt.Printf("seed: SeedFromJSON finished, created=%d\n", created)
	return created, nil
}

// FillProfilesForAll ensures every login_users row has a corresponding users profile
// and fills simple mock data (first/last/birthday/level) when missing.
// SeedProfilesFromJSON reads a JSON file with profile fields and creates or updates
// corresponding rows in the `users` table for matching `login_users` rows (by username).
func SeedProfilesFromJSON(path string) (int, error) {
	fmt.Printf("seed: SeedProfilesFromJSON start path=%s\n", path)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}

	type profileSeed struct {
		Username       string `json:"username"`
		FirstName      string `json:"first_name"`
		LastName       string `json:"last_name"`
		Birthday       string `json:"birthday"`
		ProfilePicture string `json:"profile_picture"`
		Nickname       string `json:"nickname"`
		AboutMe        string `json:"about_me"`
	}

	var profiles []profileSeed
	if err := json.Unmarshal(b, &profiles); err != nil {
		return 0, err
	}

	dbPath := "pkg/db/social_network.db"
	if err := database.Init(dbPath); err != nil {
		return 0, fmt.Errorf("init db: %w", err)
	}
	fmt.Println("seed: SeedProfilesFromJSON database initialized")

	applied := 0
	for _, p := range profiles {
		fmt.Printf("seed: processing username=%s\n", p.Username)
		var userID int
		row := database.DB.QueryRowContext(context.Background(), "SELECT id FROM login_users WHERE username = ?;", p.Username)
		if err := row.Scan(&userID); err != nil {
			fmt.Printf("seed: no login_users row for username=%s: %v\n", p.Username, err)
			// no matching login user; skip
			continue
		}
		fmt.Printf("seed: found userID=%d for username=%s\n", userID, p.Username)

		// parse birthday
		var bday *time.Time
		if p.Birthday != "" {
			if tb, perr := time.Parse("2006-01-02", p.Birthday); perr == nil {
				bday = &tb
			}
		}

		// prepare profile input
		in := models.UserProfileInput{
			ID:        userID,
			FirstName: p.FirstName,
			LastName:  p.LastName,
			Birthday:  bday,
			Level:     "basic",
		}
		if p.ProfilePicture != "" {
			in.ProfilePicture = &p.ProfilePicture
		}

		// Determine whether a users row exists using a light-weight check
		var exists int
		err := database.DB.QueryRowContext(context.Background(), "SELECT 1 FROM users WHERE id = ?;", userID).Scan(&exists)
		if err != nil {
			if err == sql.ErrNoRows {
				// create
				if cerr := queries.CreateUserProfile(context.Background(), database.DB, in); cerr == nil {
					applied++
					fmt.Printf("seed: created profile for id=%d\n", userID)
				} else {
					fmt.Printf("seed: create profile failed for id=%d: %v\n", userID, cerr)
				}
			} else {
				// other error: skip
				fmt.Printf("seed: skipping id=%d due to users table error: %v\n", userID, err)
				continue
			}
		} else {
			// update existing
			if uerr := queries.UpdateUserProfile(context.Background(), database.DB, in); uerr == nil {
				applied++
				fmt.Printf("seed: updated profile for id=%d\n", userID)
			} else {
				fmt.Printf("seed: update profile failed for id=%d: %v\n", userID, uerr)
			}
		}
	}

	return applied, nil
}

// SeedFollowersFromJSON reads follower relationships from JSON and inserts them
// into the `followers` table. JSON entries must use usernames for `follower` and
// `followed` and a `status` of either "accepted" or "pending".
func SeedFollowersFromJSON(path string) (int, error) {
	fmt.Printf("seed: SeedFollowersFromJSON start path=%s\n", path)
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return 0, err
	}

	type followSeed struct {
		Follower string `json:"follower"`
		Followed string `json:"followed"`
		Status   string `json:"status"`
	}

	var items []followSeed
	if err := json.Unmarshal(b, &items); err != nil {
		return 0, err
	}

	dbPath := "pkg/db/social_network.db"
	if err := database.Init(dbPath); err != nil {
		return 0, fmt.Errorf("init db: %w", err)
	}
	fmt.Println("seed: SeedFollowersFromJSON database initialized")

	applied := 0
	skippedMissingUser := 0
	skippedExisting := 0
	skippedError := 0

	for _, it := range items {
		// Only allow accepted or pending for now
		if it.Status != "accepted" && it.Status != "pending" {
			fmt.Printf("seed: skipping relationship with unsupported status=%s\n", it.Status)
			continue
		}

		var fid, tid int
		if err := database.DB.QueryRowContext(context.Background(), "SELECT id FROM login_users WHERE username = ?;", it.Follower).Scan(&fid); err != nil {
			log.Printf("seed: follower username not found: %s (err=%v)", it.Follower, err)
			skippedMissingUser++
			continue
		}
		if err := database.DB.QueryRowContext(context.Background(), "SELECT id FROM login_users WHERE username = ?;", it.Followed).Scan(&tid); err != nil {
			log.Printf("seed: followed username not found: %s (err=%v)", it.Followed, err)
			skippedMissingUser++
			continue
		}
		log.Printf("seed: mapping usernames %s->%s to ids %d->%d", it.Follower, it.Followed, fid, tid)
		if fid == tid {
			fmt.Printf("seed: skipping self-follow for username=%s\n", it.Follower)
			continue
		}

		// check existing
		var one int
		err := database.DB.QueryRowContext(context.Background(), "SELECT 1 FROM followers WHERE follower_id = ? AND followed_id = ?;", fid, tid).Scan(&one)
		if err == nil {
			log.Printf("seed: relationship already exists %d -> %d", fid, tid)
			skippedExisting++
			continue
		}
		if err != sql.ErrNoRows {
			log.Printf("seed: checking existing relationship failed: %v", err)
			skippedError++
			continue
		}

		req := models.FollowRequest{FollowerID: fid, FollowedID: tid, Status: it.Status}
		if err := queries.CreateFollow(context.Background(), database.DB, req, it.Status); err != nil {
			log.Printf("seed: failed to create follow %d->%d: %v", fid, tid, err)
			skippedError++
			continue
		}
		applied++
		log.Printf("seed: created follow %d->%d status=%s", fid, tid, it.Status)
	}

	log.Printf("seed: followers applied=%d skippedMissingUser=%d skippedExisting=%d skippedError=%d", applied, skippedMissingUser, skippedExisting, skippedError)
	if err := queries.RebuildAllFollowCounts(context.Background(), database.DB); err != nil {
		log.Printf("seed: warning - failed to rebuild follow counters: %v", err)
	}
	return applied, nil
}

// SeedAll runs all available seeders in the correct order: signup, profiles, then followers.
// It accepts paths to the three JSON files and returns the number of users, profiles
// and follower relationships applied (in that order) or an error.
func SeedAll(signupPath, profilesPath, followersPath string) (int, int, int, error) {
	fmt.Printf("seed: SeedAll start signup=%s profiles=%s followers=%s\n", signupPath, profilesPath, followersPath)

	usersCreated, err := SeedFromJSON(signupPath)
	if err != nil {
		return usersCreated, 0, 0, fmt.Errorf("seed users failed: %w", err)
	}

	profilesCreated, err := SeedProfilesFromJSON(profilesPath)
	if err != nil {
		return usersCreated, profilesCreated, 0, fmt.Errorf("seed profiles failed: %w", err)
	}

	followersCreated, err := SeedFollowersFromJSON(followersPath)
	if err != nil {
		return usersCreated, profilesCreated, followersCreated, fmt.Errorf("seed followers failed: %w", err)
	}

	fmt.Printf("seed: SeedAll finished users=%d profiles=%d followers=%d\n", usersCreated, profilesCreated, followersCreated)
	return usersCreated, profilesCreated, followersCreated, nil
}

// DebugDiscover runs DiscoverUsers for the given user id multiple times and
// prints the JSON results. This is a helper for manual checks and testing.
// It intentionally lives in the seed package so we can re-use existing DB init
// logic and keep tooling centralized.
func DebugDiscover(userID, iterations, limit int) error {
	if iterations <= 0 {
		iterations = 1
	}
	if limit <= 0 {
		limit = 5
	}

	dbPath := "pkg/db/social_network.db"
	if err := database.Init(dbPath); err != nil {
		return fmt.Errorf("init db: %w", err)
	}
	defer database.DB.Close()

	for i := 0; i < iterations; i++ {
		users, err := queries.DiscoverUsers(context.Background(), database.DB, userID, limit)
		if err != nil {
			fmt.Printf("DiscoverUsers error on run %d: %v\n", i+1, err)
			continue
		}
		b, _ := json.MarshalIndent(users, "", "  ")
		fmt.Printf("Discover run %d:\n%s\n", i+1, string(b))
	}

	return nil
}
