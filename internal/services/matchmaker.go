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
	RDB     *redis.Client
	Clients map[string]*websocket.Conn
}

func NewMatchmakerService(rdb *redis.Client, clients map[string]*websocket.Conn) *MatchmakerService {
	return &MatchmakerService{
		RDB:     rdb,
		Clients: clients,
	}
}

func (m *MatchmakerService) HandleFindMatch(nickname string) error {
	ctx := context.Background()

	// add nickname to match queue
	added, err := m.RDB.SAdd(ctx, "match_queue", nickname).Result()
	if err != nil {
		return err
	}
	if added == 0 {
		return errors.New("already in queue")
	}

	logger.Info("Added to match queue:", nickname)

	// check if we can find a match
	players, err := m.RDB.SMembers(ctx, "match_queue").Result()
	if err != nil {
		return err
	}

	if len(players) < 2 {
		return nil // waiting
	}

	// pick two random players
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(players), func(i, j int) {
		players[i], players[j] = players[j], players[i]
	})
	p1, p2 := players[0], players[1]

	// remove them from the queue
	_, _ = m.RDB.SRem(ctx, "match_queue", p1, p2).Result()

	// X and O symbols
	symbols := []string{"X", "O"}
	r.Shuffle(2, func(i, j int) { symbols[i], symbols[j] = symbols[j], symbols[i] })

	// notify players
	m.sendMatchFound(p1, p2, symbols[0])
	m.sendMatchFound(p2, p1, symbols[1])

	return nil
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
