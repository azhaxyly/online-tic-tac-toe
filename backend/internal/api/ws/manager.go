package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"tictactoe/internal/logger"
	"tictactoe/internal/services"
	"time"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type WSManager struct {
	mu          sync.RWMutex
	clients     map[string]*websocket.Conn
	redis       *redis.Client
	matchmaker  *services.MatchmakingService
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
		case "request_rematch":
			game, ok := m.gameManager.GetGame(nickname)
			if !ok || !game.IsFinished {
				break
			}

			opponent := ""
			if nickname == game.PlayerX {
				opponent = game.PlayerO
			} else {
				opponent = game.PlayerX
			}

			m.mu.RLock()
			opponentConn, ok := m.clients[opponent]
			m.mu.RUnlock()

			if ok {
				err := opponentConn.WriteJSON(map[string]interface{}{
					"type": "rematch_requested",
				})
				if err != nil {
					logger.Warn("failed to send rematch_requested:", err)
				}

				m.mu.Lock()
				defer m.mu.Unlock()

				if game.RematchTimer == nil {
					game.RematchTimer = time.AfterFunc(15*time.Second, func() {
						m.mu.RLock()
						defer m.mu.RUnlock()

						connSelf, okSelf := m.clients[nickname]
						connOpponent, okOpponent := m.clients[opponent]
						if !okSelf || !okOpponent {
							return
						}

						game, ok := m.gameManager.GetGame(nickname)
						if !ok {
							return
						}

						if !(game.PlayAgainX && game.PlayAgainO) {
							logger.Info("Rematch timeout: players didn't both accept")

							_ = connSelf.WriteJSON(map[string]interface{}{
								"type": "rematch_declined",
							})
							_ = connOpponent.WriteJSON(map[string]interface{}{
								"type": "rematch_declined",
							})

							m.gameManager.FinishGame(m.redis, nickname)
						}
					})
				}
			} else {
				logger.Warn("Opponent not found for rematch:", opponent)
			}
		case "accept_rematch":
			msgSelf, msgOpponent, err := m.gameManager.HandlePlayAgain(nickname)
			if err != nil {
				logger.Warn("PlayAgain error:", err)
				_ = conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
				break
			}

			// Если пока что только один игрок нажал "Play Again" — ничего не делаем
			if msgSelf == nil || msgOpponent == nil {
				break
			}

			// Оба игрока нажали — начинаем новую игру
			game, ok := m.gameManager.GetGame(nickname)
			if !ok {
				break
			}

			m.mu.Lock()
			// Останавливаем таймер ожидания реванша
			if game.RematchTimer != nil {
				game.RematchTimer.Stop()
				game.RematchTimer = nil
			}
			m.mu.Unlock()

			// Теперь отправляем обоим новое сообщение о старте реванша
			player1 := msgSelf["opponent"].(string)
			player2 := msgOpponent["opponent"].(string)

			m.mu.RLock()
			conn1, ok1 := m.clients[player1]
			conn2, ok2 := m.clients[player2]
			m.mu.RUnlock()

			if ok1 {
				_ = conn1.WriteJSON(map[string]interface{}{
					"type":     "rematch",
					"symbol":   msgSelf["symbol"],
					"opponent": msgSelf["opponent"],
				})
			}

			if ok2 {
				_ = conn2.WriteJSON(map[string]interface{}{
					"type":     "rematch",
					"symbol":   msgOpponent["symbol"],
					"opponent": msgOpponent["opponent"],
				})
			}

		case "decline_rematch":
			game, ok := m.gameManager.GetGame(nickname)
			if !ok {
				break
			}

			opponent := ""
			if nickname == game.PlayerX {
				opponent = game.PlayerO
			} else {
				opponent = game.PlayerX
			}

			m.mu.RLock()
			opponentConn, ok := m.clients[opponent]
			m.mu.RUnlock()

			if ok {
				_ = opponentConn.WriteJSON(map[string]interface{}{
					"type": "rematch_declined",
				})
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

				// m.gameManager.FinishGame(m.redis, nickname)
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
