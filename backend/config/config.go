package config

import (
	"os"

	"tictactoe/internal/logger"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort  string
	PostgresDSN string
	RedisAddr   string
	RedisPass   string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		logger.Warn(".env file not found or failed to load (fallback to OS env)")
	}

	return &Config{
		ServerPort:  os.Getenv("SERVER_PORT"),
		PostgresDSN: mustGet("POSTGRES_DSN"),
		RedisAddr:   mustGet("REDIS_ADDR"),
		RedisPass:   os.Getenv("REDIS_PASS"),
	}
}

func mustGet(key string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	logger.Error("Missing required environment variable:", key)
	panic("Missing env: " + key)
}
