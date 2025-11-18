package services

import (
	"context"
	"encoding/json"
	"time"

	"tictactoe/internal/logger"
	"tictactoe/internal/models"
	"tictactoe/internal/store"

	"github.com/redis/go-redis/v9"
)

type LeaderboardService struct {
	RDB   *redis.Client
	Store *store.UserStore
}

func NewLeaderboardService(rdb *redis.Client, store *store.UserStore) *LeaderboardService {
	return &LeaderboardService{
		RDB:   rdb,
		Store: store,
	}
}

func (s *LeaderboardService) GetLeaderboard() ([]models.LeaderboardEntry, error) {
	ctx := context.Background()
	cacheKey := "leaderboard:top10"

	// 1. Попытка получить из Redis (Кеш)
	val, err := s.RDB.Get(ctx, cacheKey).Result()
	if err == nil {
		// Если данные есть в кеше -> десериализуем и отдаем
		var cachedData []models.LeaderboardEntry
		if err := json.Unmarshal([]byte(val), &cachedData); err == nil {
			return cachedData, nil
		}
		logger.Warn("Failed to unmarshal leaderboard cache:", err)
	} else if err != redis.Nil {
		logger.Error("Redis error:", err)
	}

	// 2. Если кеша нет или ошибка -> идем в БД (Postgres)
	users, err := s.Store.GetTopUsers(10)
	if err != nil {
		return nil, err
	}

	// 3. Сохраняем результат в Redis на 60 секунд (TTL)
	// Маршалим в JSON
	data, err := json.Marshal(users)
	if err == nil {
		// SetNX не нужен, просто Set с таймером
		err = s.RDB.Set(ctx, cacheKey, data, 60*time.Second).Err()
		if err != nil {
			logger.Error("Failed to update leaderboard cache:", err)
		}
	}

	return users, nil
}
