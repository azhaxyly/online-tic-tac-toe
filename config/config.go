package config

import (
	"os"
)

type Config struct {
	ServerPort  string
	PostgresDSN string
	RedisAddr   string
	RedisPass   string
}

func Load() *Config {
	return &Config{
		ServerPort:  getEnv("SERVER_PORT", "8080"),
		PostgresDSN: getEnv("POSTGRES_DSN", "postgres://user:pass@localhost:5432/tictactoe?sslmode=disable"),
		RedisAddr:   getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPass:   getEnv("REDIS_PASS", ""),
	}
}

func getEnv(key string, fallback string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return fallback
}
