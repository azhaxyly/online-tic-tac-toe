package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"tictactoe/internal/logger"
	"tictactoe/internal/services"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type WSManager struct {
	mu          sync.RWMutex
	clients     map[string]*websocket.Conn
	redis       *redis.Client
	matchmaker  *services.MatchmakerService
	gameManager *services.GameManager
}

func NewManager(rdb *redis.Client) *WSManager {
	clients := make(map[string]*websocket.Conn)
	gameManager := services.NewGameManager()
	match := services.NewMatchmakerService(rdb, clients, gameManager)

	return &WSManager{
		clients:     clients,
		redis:       rdb,
		matchmaker:  match,
		gameManager: gameManager,
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (m *WSManager) HandleConnection(w http.ResponseWriter, r *http.Request) {
	sessionIDCookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "unauthorized (no session_id)", http.StatusUnauthorized)
		return
	}
	sessionID := sessionIDCookie.Value

	ctx := context.Background()
	nickname, err := m.redis.Get(ctx, "session:"+sessionID).Result()
	if err != nil {
		logger.Warn("No nickname in Redis for session:", sessionID)
		http.Error(w, "unauthorized (session not found)", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Error("WebSocket upgrade failed:", err)
		return
	}
	logger.Info("WebSocket connected:", nickname)

	m.mu.Lock()
	m.clients[nickname] = conn
	m.mu.Unlock()

	if err := m.redis.Incr(ctx, "online_users").Err(); err != nil {
		logger.Warn("Failed to increment online users:", err)
	}

	defer func() {
		conn.Close()
		m.mu.Lock()
		delete(m.clients, nickname)
		m.mu.Unlock()

		m.matchmaker.HandleDisconnect(nickname)

		count, err := m.redis.Decr(ctx, "online_users").Result()
		if err != nil {
			logger.Warn("failed to decrement online_users:", err)
		} else if count < 0 {
			_ = m.redis.Set(ctx, "online_users", 0, 0).Err()
		}

		logger.Info("WebSocket disconnected:", nickname)
	}()

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}

		msgType, ok := msg["type"].(string)
		if !ok {
			continue
		}

		switch msgType {
		case "find_match":
			err := m.matchmaker.HandleFindMatch(nickname)
			if err != nil {
				logger.Warn("Matchmaking error:", err)
				_ = conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
			}
		case "cancel_match":
			err := m.matchmaker.HandleCancelMatch(nickname)
			if err != nil {
				logger.Warn("Cancel match error:", err)
				_ = conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
			}
		case "move":
			cell, ok := intFrom(msg["cell"])
			if !ok {
				conn.WriteJSON(map[string]string{"type": "error", "message": "invalid cell"})
				break
			}

			moveMsg, resultMsg, err := m.gameManager.HandleMove(nickname, cell)
			if err != nil {
				conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
				break
			}

			m.sendToGame(nickname, moveMsg)
			if resultMsg != nil {
				m.sendToGame(nickname, resultMsg)

				m.gameManager.FinishGame(m.redis, nickname)
			}
		}
	}
}

func (m *WSManager) sendToGame(sender string, msg any) {
	game, ok := m.gameManager.GetGame(sender)
	if !ok {
		return
	}

	players := []string{game.PlayerX, game.PlayerO}
	for _, p := range players {
		if conn, ok := m.clients[p]; ok {
			_ = conn.WriteJSON(msg)
		}
	}
}

func intFrom(v interface{}) (int, bool) {
	f, ok := v.(float64)
	return int(f), ok
}
