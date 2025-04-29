package ws

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"tictactoe/internal/logger"
	"tictactoe/internal/services"

	"sync"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type WSManager struct {
	clients     sync.Map // map[string]*websocket.Conn
	redis       *redis.Client
	matchmaker  *services.MatchmakingService
	gameManager *services.GameManager
}

func NewManager(rdb *redis.Client) *WSManager {
	gameManager := services.NewGameManager()
	manager := &WSManager{
		redis:       rdb,
		gameManager: gameManager,
	}
	manager.matchmaker = services.NewMatchmakerService(rdb, &manager.clients, gameManager)
	return manager
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
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

	m.clients.Store(nickname, conn)

	if err := m.redis.Incr(ctx, "online_users").Err(); err != nil {
		logger.Warn("Failed to increment online_users:", err)
	}

	defer func() {
		conn.Close()
		m.clients.Delete(nickname)

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
			if err := m.matchmaker.HandleFindMatch(nickname); err != nil {
				logger.Warn("Matchmaking error:", err)
				_ = conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
			}

		case "cancel_match":
			if err := m.matchmaker.HandleCancelMatch(nickname); err != nil {
				logger.Warn("Cancel match error:", err)
				_ = conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
			}

		case "request_rematch":
			game, ok := m.gameManager.GetGame(nickname)
			if !ok || !game.IsFinished {
				break
			}
			opponent := game.PlayerO
			if nickname == game.PlayerO {
				opponent = game.PlayerX
			}

			if nickname == game.PlayerX {
				game.PlayAgainX = true
			} else {
				game.PlayAgainO = true
			}
			if game.RematchTimer == nil {
				game.RematchTimer = time.AfterFunc(15*time.Second, func() {
					if !game.PlayAgainX || !game.PlayAgainO {
						if c, ok := m.clients.Load(nickname); ok {
							c.(*websocket.Conn).WriteJSON(map[string]interface{}{"type": "rematch_declined"})
						}
						if c, ok := m.clients.Load(opponent); ok {
							c.(*websocket.Conn).WriteJSON(map[string]interface{}{"type": "rematch_declined"})
						}
						m.gameManager.FinishGame(m.redis, nickname)
					}
				})
			}
			if c, ok := m.clients.Load(opponent); ok {
				c.(*websocket.Conn).WriteJSON(map[string]interface{}{
					"type":     "rematch_requested",
					"opponent": nickname,
				})
			}

		case "accept_rematch":
			msg1, msg2, err := m.gameManager.HandlePlayAgain(nickname)
			if err != nil {
				logger.Warn("PlayAgain error:", err)
				_ = conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
				break
			}
			if msg1 == nil || msg2 == nil {
				break
			}
			if game, ok := m.gameManager.GetGame(nickname); ok && game.RematchTimer != nil {
				game.RematchTimer.Stop()
				game.RematchTimer = nil
			}
			game, _ := m.gameManager.GetGame(nickname)
			var selfMsg, oppMsg map[string]interface{}
			if nickname == game.PlayerX {
				selfMsg, oppMsg = msg1, msg2
			} else {
				selfMsg, oppMsg = msg2, msg1
			}
			if c, ok := m.clients.Load(nickname); ok {
				c.(*websocket.Conn).WriteJSON(selfMsg)
			}
			opponent := game.PlayerO
			if nickname == game.PlayerO {
				opponent = game.PlayerX
			}
			if c, ok := m.clients.Load(opponent); ok {
				c.(*websocket.Conn).WriteJSON(oppMsg)
			}

		case "decline_rematch":
			game, ok := m.gameManager.GetGame(nickname)
			if !ok {
				break
			}
			for _, p := range []string{game.PlayerX, game.PlayerO} {
				if c, ok := m.clients.Load(p); ok {
					c.(*websocket.Conn).WriteJSON(map[string]interface{}{
						"type": "rematch_declined",
					})
				}
			}

		case "rejoin_match":
			game, ok := m.gameManager.GetGame(nickname)
			if !ok {
				_ = conn.WriteJSON(map[string]string{"type": "error", "message": "no active game"})
				conn.Close()
				return
			}
			_ = conn.WriteJSON(map[string]interface{}{
				"type":       "game_state",
				"board":      game.Board,
				"turn":       game.Turn,
				"isFinished": game.IsFinished,
				"winner":     game.Winner,
			})
		case "forfeit":
			game, ok := m.gameManager.GetGame(nickname)
			if ok && !game.IsFinished {
				winner := game.PlayerO
				if nickname == game.PlayerO {
					winner = game.PlayerX
				}
				m.sendToGame(nickname, map[string]interface{}{
					"type":   "game_over",
					"result": winner,
				})
				game.IsFinished = true
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
				if game, ok := m.gameManager.GetGame(nickname); ok {
					game.IsFinished = true
				}
			} else {
				if game, ok := m.gameManager.GetGame(nickname); ok {
					boardFull := true
					for _, cellVal := range game.Board {
						if cellVal == "" {
							boardFull = false
							break
						}
					}
					if boardFull {
						drawMsg := map[string]interface{}{
							"type":   "game_over",
							"result": "draw",
						}
						m.sendToGame(nickname, drawMsg)

						game.IsFinished = true
					}
				}
			}
		}
	}
}

func (m *WSManager) sendToGame(sender string, msg any) {
	game, ok := m.gameManager.GetGame(sender)
	if !ok {
		return
	}
	for _, p := range []string{game.PlayerX, game.PlayerO} {
		if val, ok := m.clients.Load(p); ok {
			conn := val.(*websocket.Conn)
			_ = conn.WriteJSON(msg)
		}
	}
}

func intFrom(v interface{}) (int, bool) {
	f, ok := v.(float64)
	return int(f), ok
}
