package cache

import (
	"context"
	"fmt"

	"tictactoe/internal/logger"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func NewRedis(addr, password string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	if err := rdb.Ping(Ctx).Err(); err != nil {
		logger.Error("Redis ping failed:", err)
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	logger.Info("Redis connected successfully")
	return rdb, nil
}
