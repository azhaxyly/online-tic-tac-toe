package models

type User struct {
	ID           int    `json:"-"`
	Nickname     string `json:"nickname"`
	PasswordHash string `json:"-"`
	Wins         int    `json:"wins"`
	Losses       int    `json:"losses"`
	Draws        int    `json:"draws"`
}

type LeaderboardEntry struct {
	Nickname string `json:"nickname"`
	Wins     int    `json:"wins"`
	Losses   int    `json:"losses"`
	Draws    int    `json:"draws"`
}
