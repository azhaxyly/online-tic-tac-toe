package store

import (
	"database/sql"
	"errors"
	"fmt"

	"tictactoe/internal/models"
)

type UserStore struct {
	DB *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{DB: db}
}

func (s *UserStore) CreateUser(nickname, passwordHash string) (*models.User, error) {
	var id int
	err := s.DB.QueryRow(`
		INSERT INTO users (nickname, password_hash) VALUES ($1, $2)
		RETURNING id
	`, nickname, passwordHash).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}
	return &models.User{ID: id, Nickname: nickname}, nil
}

func (s *UserStore) GetUserByNickname(nickname string) (*models.User, string, error) {
	user := &models.User{Nickname: nickname}
	var passwordHash string

	err := s.DB.QueryRow(`
		SELECT id, password_hash FROM users WHERE nickname = $1
	`, nickname).Scan(&user.ID, &passwordHash)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, "", fmt.Errorf("get user: %w", err)
	}

	return user, passwordHash, nil
}
