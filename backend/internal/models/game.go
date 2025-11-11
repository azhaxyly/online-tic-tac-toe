package models

import "time"

type Game struct {
	PlayerX       string
	PlayerO       string
	Turn          string
	Board         [9]string
	IsFinished    bool
	Winner        string
	PlayAgainX    bool
	PlayAgainO    bool
	RematchTimer  *time.Timer
	IsBotGame     bool          // NEW: флаг игры с ботом
	BotDifficulty BotDifficulty // NEW: сложность бота
	BotSymbol     string        // NEW: символ бота (X или O)
}
