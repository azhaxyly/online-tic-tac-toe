package store

import (
	"database/sql"
	"fmt"

	"tictactoe/internal/logger"

	_ "github.com/lib/pq"
)

func NewPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Error("Failed to open Postgres:", err)
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	if err = db.Ping(); err != nil {
		logger.Error("Failed to ping Postgres:", err)
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	logger.Info("PostgreSQL connected successfully")
	return db, nil
}
