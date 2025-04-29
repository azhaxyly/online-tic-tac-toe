package services

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"tictactoe/internal/logger"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type MatchmakingService struct {
	RDB         *redis.Client
	Clients     *sync.Map
	GameManager *GameManager
}

func NewMatchmakerService(rdb *redis.Client, clientsMap *sync.Map, gm *GameManager) *MatchmakingService {
	return &MatchmakingService{
		RDB:         rdb,
		Clients:     clientsMap,
		GameManager: gm,
	}
}

func (m *MatchmakingService) HandleFindMatch(nickname string) error {
	ctx := context.Background()

	added, err := m.RDB.SAdd(ctx, "match_queue", nickname).Result()
	if err != nil {
		return err
	}
	if added == 0 {
		return errors.New("already in queue")
	}

	logger.Info("Added to match queue:", nickname)

	if c, ok := m.Clients.Load(nickname); ok {
		conn := c.(*websocket.Conn)
		_ = conn.WriteJSON(map[string]interface{}{"type": "searching"})
	}

	players, err := m.RDB.SMembers(ctx, "match_queue").Result()
	if err != nil {
		return err
	}
	if len(players) < 2 {
		return nil
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})
	p1, p2 := players[0], players[1]
	_, _ = m.RDB.SRem(ctx, "match_queue", p1, p2).Result()

	symbols := []string{"X", "O"}
	r.Shuffle(2, func(i, j int) { symbols[i], symbols[j] = symbols[j], symbols[i] })

	m.sendMatchFound(p1, p2, symbols[0])
	m.sendMatchFound(p2, p1, symbols[1])

	m.GameManager.CreateGame(p1, p2, symbols[0], symbols[1])

	if err := m.RDB.Incr(ctx, "active_games").Err(); err != nil {
		logger.Warn("failed to increment active_games:", err)
	}

	return nil
}

func (m *MatchmakingService) HandleCancelMatch(nickname string) error {
	ctx := context.Background()
	removed, err := m.RDB.SRem(ctx, "match_queue", nickname).Result()
	if err != nil {
		return err
	}
	if removed == 0 {
		return errors.New("not in queue")
	}

	if c, ok := m.Clients.Load(nickname); ok {
		conn := c.(*websocket.Conn)
		_ = conn.WriteJSON(map[string]interface{}{"type": "match_cancelled"})
	}
	logger.Info("Removed from match queue:", nickname)
	return nil
}

func (m *MatchmakingService) HandleDisconnect(nickname string) {
	ctx := context.Background()
	if _, err := m.RDB.SRem(ctx, "match_queue", nickname).Result(); err != nil {
		logger.Warn("failed to remove from match_queue:", err)
	}

	game, ok := m.GameManager.GetGame(nickname)
	if !ok {
		return
	}
	opponent := game.PlayerO
	if nickname == game.PlayerO {
		opponent = game.PlayerX
	}

	if c, ok := m.Clients.Load(opponent); ok {
		conn := c.(*websocket.Conn)
		_ = conn.WriteJSON(map[string]interface{}{"type": "opponent_left"})
	}

	m.GameManager.FinishGame(m.RDB, nickname)
}

func (m *MatchmakingService) sendMatchFound(player, opponent, symbol string) {
	if c, ok := m.Clients.Load(player); ok {
		conn := c.(*websocket.Conn)
		msg := map[string]interface{}{
			"type":     "match_found",
			"symbol":   symbol,
			"opponent": opponent,
		}
		_ = conn.WriteJSON(msg)
	}
}
