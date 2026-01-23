package seed

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"

	queries "backend/pkg/db/queries"
	database "backend/pkg/db/sqlite"
)

type SeedUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// SeedFromJSON reads the given JSON file and creates users via queries.SignUp.
// It returns the number of users created or an error.
func SeedFromJSON(path string) (int, error) {
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

	created := 0
	for _, u := range users {
		input := queries.SignUpInput{Email: u.Email, Username: u.Username, Password: u.Password}
		if err := queries.SignUp(context.Background(), database.DB, input); err != nil {
			switch err {
			case queries.ErrEmailTaken, queries.ErrUsernameTaken:
				// skip duplicates
				continue
			default:
				// non-fatal: continue seeding other users
				continue
			}
		}
		created++
	}

	return created, nil
}
