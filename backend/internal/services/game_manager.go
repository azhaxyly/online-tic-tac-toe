package services

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"tictactoe/internal/logger"
	"tictactoe/internal/models"
	"time"

	"github.com/redis/go-redis/v9"
)

type GameManager struct {
	mu    sync.RWMutex
	games map[string]*models.Game
}

func NewGameManager() *GameManager {
	return &GameManager{
		games: make(map[string]*models.Game),
	}
}

func (g *GameManager) CreateGame(p1, p2, sym1, sym2 string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	playerX := p1
	playerO := p2

	if sym1 != "X" {
		playerX = p2
		playerO = p1
	}

	game := &models.Game{
		PlayerX:    playerX,
		PlayerO:    playerO,
		Turn:       "X",
		Board:      [9]string{},
		IsFinished: false,
	}

	g.games[playerX] = game
	g.games[playerO] = game
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

	winner, winningPattern := checkWin(game.Board)
	if winner != "" {
		game.IsFinished = true
		game.Winner = winner

		result := map[string]interface{}{
			"type":           "game_over",
			"result":         winner,
			"winningPattern": winningPattern,
		}

		return move, result, nil
	}

	return move, nil, nil
}

func (g *GameManager) GetGame(nickname string) (*models.Game, bool) {
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

	ctx := context.Background()

	switch game.Winner {
	case "X":
		rdb.Incr(ctx, "wins:"+game.PlayerX)
		rdb.Incr(ctx, "losses:"+game.PlayerO)
	case "O":
		rdb.Incr(ctx, "wins:"+game.PlayerO)
		rdb.Incr(ctx, "losses:"+game.PlayerX)
	case "draw":
		rdb.Incr(ctx, "draws:"+game.PlayerX)
		rdb.Incr(ctx, "draws:"+game.PlayerO)
	}

	delete(g.games, game.PlayerX)
	delete(g.games, game.PlayerO)

	count, err := rdb.Decr(ctx, "active_games").Result()
	if err != nil {
		logger.Warn("failed to decrement active_games:", err)
	} else if count < 0 {
		_ = rdb.Set(ctx, "active_games", 0, 0).Err()
	}
}

func (g *GameManager) HandlePlayAgain(nickname string) (map[string]interface{}, map[string]interface{}, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	game, ok := g.games[nickname]
	if !ok {
		return nil, nil, fmt.Errorf("no active game")
	}

	if nickname == game.PlayerX {
		game.PlayAgainX = true
	} else if nickname == game.PlayerO {
		game.PlayAgainO = true
	} else {
		return nil, nil, fmt.Errorf("not a player")
	}

	if !(game.PlayAgainX && game.PlayAgainO) {
		return nil, nil, nil
	}

	if game.RematchTimer != nil {
		game.RematchTimer.Stop()
		game.RematchTimer = nil
	}

	players := []string{game.PlayerX, game.PlayerO}
	symbols := []string{"X", "O"}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(2, func(i, j int) { symbols[i], symbols[j] = symbols[j], symbols[i] })

	game.Board = [9]string{}
	game.IsFinished = false
	game.Turn = "X"
	game.PlayAgainX = false
	game.PlayAgainO = false
	game.Winner = ""

	if symbols[0] == "X" {
		game.PlayerX = players[0]
		game.PlayerO = players[1]
	} else {
		game.PlayerX = players[1]
		game.PlayerO = players[0]
	}

	msg1 := map[string]interface{}{
		"type":     "rematch",
		"symbol":   "X",
		"opponent": game.PlayerO,
	}
	msg2 := map[string]interface{}{
		"type":     "rematch",
		"symbol":   "O",
		"opponent": game.PlayerX,
	}

	return msg1, msg2, nil
}

func opposite(s string) string {
	if s == "X" {
		return "O"
	}
	return "X"
}

func checkWin(board [9]string) (string, []int) {
	winPatterns := [][]int{
		{0, 1, 2}, {3, 4, 5}, {6, 7, 8},
		{0, 3, 6}, {1, 4, 7}, {2, 5, 8},
		{0, 4, 8}, {2, 4, 6},
	}

	for _, pattern := range winPatterns {
		a, b, c := pattern[0], pattern[1], pattern[2]
		if board[a] != "" && board[a] == board[b] && board[b] == board[c] {
			return board[a], pattern
		}
	}

	return "", nil
}
