package services

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"tictactoe/internal/logger"
	"tictactoe/internal/models"
	"tictactoe/internal/store"

	"github.com/redis/go-redis/v9"
)

type GameManager struct {
	mu        sync.RWMutex
	games     map[string]*models.Game
	userStore *store.UserStore
}

func NewGameManager(userStore *store.UserStore) *GameManager {
	return &GameManager{
		games:     make(map[string]*models.Game),
		userStore: userStore,
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
		PlayerX:      playerX,
		PlayerO:      playerO,
		Turn:         "X",
		Board:        [9]string{},
		IsFinished:   false,
		LastActivity: time.Now(),
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

	game.LastActivity = time.Now()
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
		g.updateElo(game.PlayerX, game.PlayerO, 1.0)
	case "O":
		rdb.Incr(ctx, "wins:"+game.PlayerO)
		rdb.Incr(ctx, "losses:"+game.PlayerX)
		g.updateElo(game.PlayerO, game.PlayerX, 1.0)
	case "draw":
		rdb.Incr(ctx, "draws:"+game.PlayerX)
		rdb.Incr(ctx, "draws:"+game.PlayerO)
		g.updateElo(game.PlayerX, game.PlayerO, 0.5)
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

func (g *GameManager) updateElo(playerA, playerB string, scoreA float64) {
	// Получаем текущие рейтинги
	userA, _, errA := g.userStore.GetUserByNickname(playerA)
	userB, _, errB := g.userStore.GetUserByNickname(playerB)

	if errA != nil || errB != nil {
		logger.Error("Failed to get users for ELO update:", errA, errB)
		return
	}

	// Расчет изменения рейтинга
	kFactor := 32
	expectedA := 1.0 / (1.0 + math.Pow(10, float64(userB.EloRating-userA.EloRating)/400.0))
	changeA := int(float64(kFactor) * (scoreA - expectedA))

	// Обновляем статистику в БД
	resultA := "draw"
	resultB := "draw"
	if scoreA == 1.0 {
		resultA = "win"
		resultB = "loss"
	} else if scoreA == 0.0 {
		resultA = "loss"
		resultB = "win"
	}

	if err := g.userStore.UpdateUserStats(playerA, changeA, resultA); err != nil {
		logger.Error("Failed to update stats for", playerA, ":", err)
	}
	// Для второго игрока изменение противоположное
	if err := g.userStore.UpdateUserStats(playerB, -changeA, resultB); err != nil {
		logger.Error("Failed to update stats for", playerB, ":", err)
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
		{0, 1, 2},
		{3, 4, 5},
		{6, 7, 8},
		{0, 3, 6},
		{1, 4, 7},
		{2, 5, 8},
		{0, 4, 8},
		{2, 4, 6},
	}

	for _, pattern := range winPatterns {
		a, b, c := pattern[0], pattern[1], pattern[2]
		if board[a] != "" && board[a] == board[b] && board[b] == board[c] {
			return board[a], pattern
		}
	}

	return "", nil
}

func (g *GameManager) CreateBotGame(player string, difficulty models.BotDifficulty) string {
	g.mu.Lock()
	defer g.mu.Unlock()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	playerSymbol := "X"
	botSymbol := "O"
	if r.Intn(2) == 1 {
		playerSymbol = "O"
		botSymbol = "X"
	}
	botName := fmt.Sprintf("Bot_%s", difficulty)

	var playerX, playerO string
	if playerSymbol == "X" {
		playerX = player
		playerO = botName
	} else {
		playerX = botName
		playerO = player
	}

	game := &models.Game{
		PlayerX:       playerX,
		PlayerO:       playerO,
		Turn:          "X",
		Board:         [9]string{},
		IsFinished:    false,
		IsBotGame:     true,
		BotDifficulty: difficulty,
		BotSymbol:     botSymbol,
		LastActivity:  time.Now(),
	}

	g.games[player] = game
	g.games[botName] = game

	return playerSymbol
}

func (g *GameManager) IsBotTurn(nickname string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	game, ok := g.games[nickname]
	if !ok || !game.IsBotGame {
		return false
	}

	if game.Turn == game.BotSymbol {
		return true
	}

	return false
}

func (g *GameManager) StartCleaner(rdb *redis.Client) {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for range ticker.C {
			g.cleanupAndSync(rdb)
		}
	}()
}

func (g *GameManager) cleanupAndSync(rdb *redis.Client) {
	g.mu.Lock()
	defer g.mu.Unlock()

	ctx := context.Background()
	uniqueGames := make(map[*models.Game]bool)
	keysToDelete := []string{}

	for nickname, game := range g.games {
		if game.IsFinished || time.Since(game.LastActivity) > 2*time.Minute {
			keysToDelete = append(keysToDelete, nickname)
		} else {
			uniqueGames[game] = true
		}
	}

	for _, key := range keysToDelete {
		delete(g.games, key)
	}

	realCount := len(uniqueGames)

	err := rdb.Set(ctx, "active_games", realCount, 0).Err()
	if err != nil {
		logger.Error("Failed to sync active_games:", err)
	}
}
