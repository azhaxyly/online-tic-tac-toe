package store

import (
	"fmt"
	"tictactoe/internal/models"
)

func (s *UserStore) GetTopUsers(limit int) ([]models.LeaderboardEntry, error) {
	rows, err := s.DB.Query(`
        SELECT nickname, wins, losses, draws, elo_rating 
        FROM users 
        ORDER BY elo_rating DESC 
        LIMIT $1
    `, limit)
	if err != nil {
		return nil, fmt.Errorf("query leaderboard: %w", err)
	}
	defer rows.Close()

	var users []models.LeaderboardEntry
	for rows.Next() {
		var u models.LeaderboardEntry
		if err := rows.Scan(&u.Nickname, &u.Wins, &u.Losses, &u.Draws, &u.EloRating); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *UserStore) UpdateUserStats(nickname string, eloChange int, result string) error {
	query := `
		UPDATE users 
		SET elo_rating = elo_rating + $1,
			total_games = total_games + 1,
			wins = wins + CASE WHEN $2 = 'win' THEN 1 ELSE 0 END,
			losses = losses + CASE WHEN $2 = 'loss' THEN 1 ELSE 0 END,
			draws = draws + CASE WHEN $2 = 'draw' THEN 1 ELSE 0 END
		WHERE nickname = $3
	`
	_, err := s.DB.Exec(query, eloChange, result, nickname)
	if err != nil {
		return fmt.Errorf("update user stats: %w", err)
	}
	return nil
}
