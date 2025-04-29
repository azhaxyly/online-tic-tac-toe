package services

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"tictactoe/internal/models"
	"tictactoe/internal/store"

	"github.com/redis/go-redis/v9"
)

type SessionService struct {
	RDB   *redis.Client
	Store *store.SessionStore
}

func NewSessionService(rdb *redis.Client, store *store.SessionStore) *SessionService {
	return &SessionService{RDB: rdb, Store: store}
}

func (s *SessionService) GetOrCreateUser(sessionID string) (*models.User, error) {
	ctx := context.Background()

	if nickname, err := s.RDB.Get(ctx, "session:"+sessionID).Result(); err == nil {
		return &models.User{Nickname: nickname}, nil
	}

	nickname := s.generateNickname()

	user, err := s.Store.InsertUser(nickname)
	if err != nil {
		return nil, err
	}

	_ = s.RDB.Set(ctx, "session:"+sessionID, nickname, 24*time.Hour).Err()
	return user, nil
}

func (s *SessionService) generateNickname() string {
	words := []string{"Alem", "Tomorrow", "Jews", "Pups", "Pudge", "Gau", "Tajik"}
	n := rand.Intn(10000)
	adjective := words[rand.Intn(len(words))]
	return fmt.Sprintf("%s%d", adjective, n)
}
