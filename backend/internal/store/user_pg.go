package store

import (
	"fmt"
	"tictactoe/internal/models"
)

func (s *UserStore) GetTopUsers(limit int) ([]models.LeaderboardEntry, error) {
	rows, err := s.DB.Query(`
        SELECT nickname, wins, losses, draws 
        FROM users 
        ORDER BY wins DESC 
        LIMIT $1
    `, limit)
	if err != nil {
		return nil, fmt.Errorf("query leaderboard: %w", err)
	}
	defer rows.Close()

	var users []models.LeaderboardEntry
	for rows.Next() {
		var u models.LeaderboardEntry
		if err := rows.Scan(&u.Nickname, &u.Wins, &u.Losses, &u.Draws); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
