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
	mu         sync.RWMutex
	clients    map[string]*websocket.Conn
	redis      *redis.Client
	matchmaker *services.MatchmakerService
}

func NewManager(rdb *redis.Client) *WSManager {
	clients := make(map[string]*websocket.Conn)
	match := services.NewMatchmakerService(rdb, clients)

	return &WSManager{
		clients:    clients,
		redis:      rdb,
		matchmaker: match,
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

	defer func() {
		conn.Close()
		m.mu.Lock()
		delete(m.clients, nickname)
		m.mu.Unlock()
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
		}
	}
}
