package cmd

import (
	"os"
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

	port := os.Getenv("PORT")

	if port == "" {
		port = cfg.ServerPort
	}

	if port == "" {
		port = "8080"
	}

	logger.Info("Starting HTTP server on port:", port)

	if err := router.Run(":" + port); err != nil {
		logger.Error("Failed to start server:", err)
	}
}
