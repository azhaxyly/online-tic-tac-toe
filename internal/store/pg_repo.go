package store

import (
	"database/sql"
	"fmt"
	"time"

	"tictactoe/internal/logger"

	_ "github.com/lib/pq"
)

func NewPostgres(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", dsn)
		if err != nil {
			logger.Warn("Retry Postgres open:", err)
			time.Sleep(2 * time.Second)
			continue
		}
		if err = db.Ping(); err == nil {
			logger.Info("PostgreSQL connected successfully")
			return db, nil
		}
		logger.Warn("Retry Postgres ping:", err)
		time.Sleep(2 * time.Second)
	}

	logger.Error("Failed to connect to Postgres after retries:", err)
	return nil, fmt.Errorf("ping postgres after retries: %w", err)
}
