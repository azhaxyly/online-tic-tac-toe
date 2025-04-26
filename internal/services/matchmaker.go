package services

import (
	"context"
	"errors"
	"math/rand"
	"tictactoe/internal/logger"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type MatchmakerService struct {
	RDB         *redis.Client
	Clients     map[string]*websocket.Conn
	GameManager *GameManager
}

func NewMatchmakerService(rdb *redis.Client, clients map[string]*websocket.Conn, gm *GameManager) *MatchmakerService {
	return &MatchmakerService{
		RDB:         rdb,
		Clients:     clients,
		GameManager: gm,
	}
}

func (m *MatchmakerService) HandleFindMatch(nickname string) error {
	ctx := context.Background()

	added, err := m.RDB.SAdd(ctx, "match_queue", nickname).Result()
	if err != nil {
		return err
	}
	if added == 0 {
		return errors.New("already in queue")
	}

	logger.Info("Added to match queue:", nickname)

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

func (m *MatchmakerService) HandleCancelMatch(nickname string) error {
	ctx := context.Background()

	removed, err := m.RDB.SRem(ctx, "match_queue", nickname).Result()
	if err != nil {
		return err
	}

	if removed == 0 {
		return errors.New("not in queue")
	}

	conn, ok := m.Clients[nickname]
	if ok {
		msg := map[string]interface{}{
			"type": "match_cancelled",
		}
		_ = conn.WriteJSON(msg)
	}

	logger.Info("Removed from match queue:", nickname)
	return nil
}

func (m *MatchmakerService) HandleDisconnect(nickname string) {
	ctx := context.Background()

	_, err := m.RDB.SRem(ctx, "match_queue", nickname).Result()
	if err != nil {
		logger.Warn("failed to remove from match_queue:", err)
	}

	game, ok := m.GameManager.GetGame(nickname)
	if !ok {
		return
	}

	opponent := ""
	if nickname == game.PlayerX {
		opponent = game.PlayerO
	} else {
		opponent = game.PlayerX
	}

	conn, ok := m.Clients[opponent]
	if ok {
		msg := map[string]interface{}{
			"type": "opponent_left",
		}
		_ = conn.WriteJSON(msg)
	}

	m.GameManager.FinishGame(m.RDB, nickname)
}

func (m *MatchmakerService) sendMatchFound(player, opponent, symbol string) {
	conn, ok := m.Clients[player]
	if !ok {
		return
	}
	msg := map[string]interface{}{
		"type":     "match_found",
		"symbol":   symbol,
		"opponent": opponent,
	}
	_ = conn.WriteJSON(msg)
}
