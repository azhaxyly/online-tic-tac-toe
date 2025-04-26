package services

import (
	"context"
	"fmt"
	"sync"
	"tictactoe/internal/logger"

	"github.com/redis/go-redis/v9"
)

type Game struct {
	PlayerX    string
	PlayerO    string
	Board      [9]string
	Turn       string // "X" or "O"
	IsFinished bool
	Winner     string // "X", "O", "draw"
}

type GameManager struct {
	mu    sync.RWMutex
	games map[string]*Game
}

func NewGameManager() *GameManager {
	return &GameManager{
		games: make(map[string]*Game),
	}
}

func (g *GameManager) CreateGame(p1, p2, sym1, sym2 string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	var px, po string
	if sym1 == "X" {
		px, po = p1, p2
	} else {
		px, po = p2, p1
	}

	game := &Game{
		PlayerX:    px,
		PlayerO:    po,
		Turn:       "X",
		Board:      [9]string{},
		IsFinished: false,
	}

	g.games[px] = game
	g.games[po] = game
}

func (g *GameManager) HandleMove(nickname string, cell int) (map[string]interface{}, map[string]interface{}, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	game, ok := g.games[nickname]
	if !ok || game.IsFinished {
		return nil, nil, fmt.Errorf("no active game")
	}

	symbol := ""
	if nickname == game.PlayerX {
		symbol = "X"
	} else if nickname == game.PlayerO {
		symbol = "O"
	} else {
		return nil, nil, fmt.Errorf("not a player")
	}

	if game.Turn != symbol {
		return nil, nil, fmt.Errorf("not your turn")
	}
	if cell < 0 || cell > 8 || game.Board[cell] != "" {
		return nil, nil, fmt.Errorf("invalid move")
	}

	game.Board[cell] = symbol
	game.Turn = opposite(symbol)

	move := map[string]interface{}{
		"type": "move_made",
		"cell": cell,
		"by":   symbol,
	}

	if winner := checkWin(game.Board); winner != "" {
		game.IsFinished = true
		game.Winner = winner

		result := map[string]interface{}{
			"type":   "game_over",
			"result": winner,
		}

		return move, result, nil
	}

	return move, nil, nil
}

func (g *GameManager) GetGame(nickname string) (*Game, bool) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	game, ok := g.games[nickname]
	return game, ok
}

func (g *GameManager) FinishGame(rdb *redis.Client, nickname string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	game, ok := g.games[nickname]
	if !ok {
		return
	}

	delete(g.games, game.PlayerX)
	delete(g.games, game.PlayerO)

	ctx := context.Background()
	count, err := rdb.Decr(ctx, "active_games").Result()
	if err != nil {
		logger.Warn("failed to decrement active_games:", err)
	} else if count < 0 {
		_ = rdb.Set(ctx, "active_games", 0, 0).Err()
	}
}

func opposite(s string) string {
	if s == "X" {
		return "O"
	}
	return "X"
}

func checkWin(board [9]string) string {
	lines := [][3]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8},
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8},
		{0, 4, 8}, {2, 4, 6},
	}
	for _, l := range lines {
		a, b, c := l[0], l[1], l[2]
		if board[a] != "" && board[a] == board[b] && board[b] == board[c] {
			return board[a]
		}
	}
	for _, cell := range board {
		if cell == "" {
			return ""
		}
	}
	return "draw"
}
