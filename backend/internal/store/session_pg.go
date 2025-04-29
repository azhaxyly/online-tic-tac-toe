package store

import (
	"database/sql"
	"fmt"

	"tictactoe/internal/models"
)

type SessionStore struct {
	DB *sql.DB
}

func NewSessionStore(db *sql.DB) *SessionStore {
	return &SessionStore{DB: db}
}

func (s *SessionStore) InsertUser(nickname string) (*models.User, error) {
	var id int
	err := s.DB.QueryRow(`
		INSERT INTO users (nickname) VALUES ($1)
		RETURNING id
	`, nickname).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return &models.User{ID: id, Nickname: nickname}, nil
}
