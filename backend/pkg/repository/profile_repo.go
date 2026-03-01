package repository

import (
	"database/sql"
)

type ProfileRepository interface {
}

type sqliteProfileRepo struct {
	db *sql.DB
}

func NewProfileRepository(db *sql.DB) ProfileRepository {
	return &sqliteProfileRepo{db: db}
}

