package cmd

import (
	"tictactoe/config"
	"tictactoe/internal/cache"
	"tictactoe/internal/logger"
	"tictactoe/internal/store"

	"github.com/gin-gonic/gin"
)

func Run() {
	logger.Init()

	cfg := config.Load()
	logger.Info("Configuration loaded")

	db, err := store.NewPostgres(cfg.PostgresDSN)
	if err != nil {
		logger.Error("Postgres connection failed:", err)
		return
	}
	defer db.Close()
	logger.Info("Connected to PostgreSQL")

	rdb, err := cache.NewRedis(cfg.RedisAddr, cfg.RedisPass)
	if err != nil {
		logger.Error("failed to connect to Redis:", err)
		return
	}
	defer func() {
		if err := rdb.Close(); err != nil {
			logger.Error("failed to close Redis:", err)
		}
	}()
	logger.Info("Connected to Redis")

	router := gin.Default()
	logger.Info("Starting HTTP server on port:", cfg.ServerPort)

	// TODO: router.Use(), router.GET/POST, ws handler

	if err := router.Run(":" + cfg.ServerPort); err != nil {
		logger.Error("Failed to start server:", err)
	}
}
