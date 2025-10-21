package cmd

import (
	"tictactoe/config"
	"tictactoe/internal/api/http"
	"tictactoe/internal/cache"
	"tictactoe/internal/logger"
	"tictactoe/internal/services"
	"tictactoe/internal/store"
)

func Run() {
	logger.Init()

	cfg := config.Load()

	db, err := store.NewPostgres(cfg.PostgresDSN)
	if err != nil {
		logger.Error("Postgres connection failed:", err)
		return
	}
	defer db.Close()

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

	sessionStore := store.NewUserStore(db)
	sessionService := services.NewSessionService(rdb, sessionStore)

	router := http.NewRouter(sessionService)
	logger.Info("Starting HTTP server on port:", cfg.ServerPort)

	if err := router.Run(":" + cfg.ServerPort); err != nil {
		logger.Error("Failed to start server:", err)
	}
}
