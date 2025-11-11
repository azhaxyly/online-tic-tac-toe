package models

type BotDifficulty string

const (
	DifficultyEasy   BotDifficulty = "easy"
	DifficultyMedium BotDifficulty = "medium"
	DifficultyHard   BotDifficulty = "hard"
)

type BotPlayer struct {
	Difficulty BotDifficulty
	Symbol     string
}
