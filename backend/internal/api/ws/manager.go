package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"tictactoe/internal/models"
	"time"

	"tictactoe/internal/logger"
	"tictactoe/internal/services"

	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

type WSManager struct {
	clients     sync.Map
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

func (m *WSManager) HandleConnection(w http.ResponseWriter, r *http.Request, nickname string) {
	conn, err := m.upgradeConnection(w, r)
	if err != nil {
		logger.Error("WebSocket upgrade failed:", err)
		return
	}

	logger.Info("WebSocket connected:", nickname)
	m.clients.Store(nickname, conn)

	ctx := context.Background()
	_ = m.redis.Incr(ctx, "online_users").Err()

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

	m.handleMessages(conn, nickname)
}

func (m *WSManager) upgradeConnection(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (m *WSManager) handleMessages(conn *websocket.Conn, nickname string) {
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

		m.handleMessageType(conn, nickname, msgType, msg)
	}
}

func (m *WSManager) handleMessageType(conn *websocket.Conn, nickname, msgType string, msg map[string]interface{}) {
	switch msgType {
	case "find_match":
		if err := m.matchmaker.HandleFindMatch(nickname); err != nil {
			logger.Warn("Matchmaking error:", err)
			_ = conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
		}
	case "find_bot_match":
		difficulty, ok := msg["difficulty"].(string)
		if !ok {
			_ = conn.WriteJSON(map[string]string{"type": "error", "message": "invalid difficulty"})
			return
		}
		m.handleFindBotMatch(conn, nickname, models.BotDifficulty(difficulty))
	case "cancel_match":
		if err := m.matchmaker.HandleCancelMatch(nickname); err != nil {
			logger.Warn("Cancel match error:", err)
			_ = conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
		}
	case "request_rematch":
		m.handleRequestRematch(nickname)
	case "accept_rematch":
		m.handleAcceptRematch(conn, nickname)
	case "decline_rematch":
		m.handleDeclineRematch(nickname)
	case "rejoin_match":
		m.handleRejoinMatch(conn, nickname)
	case "forfeit":
		m.handleForfeit(nickname)
	case "move":
		m.handleMove(conn, nickname, msg)
	default:
		logger.Warn("Unhandled message type:", msgType)
	}
}

func (m *WSManager) handleRequestRematch(nickname string) {
	game, ok := m.gameManager.GetGame(nickname)
	if !ok || !game.IsFinished {
		return
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
}

func (m *WSManager) handleAcceptRematch(conn *websocket.Conn, nickname string) {
	msg1, msg2, err := m.gameManager.HandlePlayAgain(nickname)
	if err != nil {
		logger.Warn("PlayAgain error:", err)
		_ = conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
		return
	}
	if msg1 == nil || msg2 == nil {
		return
	}
	game, _ := m.gameManager.GetGame(nickname)
	if game.RematchTimer != nil {
		game.RematchTimer.Stop()
		game.RematchTimer = nil
	}

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
}

func (m *WSManager) handleDeclineRematch(nickname string) {
	game, ok := m.gameManager.GetGame(nickname)
	if !ok {
		return
	}
	for _, p := range []string{game.PlayerX, game.PlayerO} {
		if c, ok := m.clients.Load(p); ok {
			c.(*websocket.Conn).WriteJSON(map[string]interface{}{
				"type": "rematch_declined",
			})
		}
	}
	m.gameManager.FinishGame(m.redis, nickname)
}

func (m *WSManager) handleRejoinMatch(conn *websocket.Conn, nickname string) {
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
}

func (m *WSManager) handleForfeit(nickname string) {
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
		//m.gameManager.FinishGame(m.redis, nickname)
	}
}

func (m *WSManager) handleMove(conn *websocket.Conn, nickname string, msg map[string]interface{}) {
	cell, ok := intFrom(msg["cell"])
	if !ok {
		conn.WriteJSON(map[string]string{"type": "error", "message": "invalid cell"})
		return
	}

	moveMsg, resultMsg, err := m.gameManager.HandleMove(nickname, cell)
	if err != nil {
		conn.WriteJSON(map[string]string{"type": "error", "message": err.Error()})
		return
	}

	game, _ := m.gameManager.GetGame(nickname)

	// Для игры с ботом отправляем только игроку
	if game.IsBotGame {
		_ = conn.WriteJSON(moveMsg)

		if resultMsg != nil {
			_ = conn.WriteJSON(resultMsg)
			game.IsFinished = true
		} else {
			// Проверяем ничью
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
				_ = conn.WriteJSON(drawMsg)
				game.IsFinished = true
			} else {
				go func() {
					time.Sleep(500 * time.Millisecond)
					m.makeBotMove(nickname)
				}()
			}
		}
	} else {
		// Оригинальная логика для PvP
		m.sendToGame(nickname, moveMsg)
		if resultMsg != nil {
			m.sendToGame(nickname, resultMsg)
			game.IsFinished = true
		} else {
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

func (m *WSManager) handleFindBotMatch(conn *websocket.Conn, nickname string, difficulty models.BotDifficulty) {
	// Проверяем валидность сложности
	if difficulty != models.DifficultyEasy &&
		difficulty != models.DifficultyMedium &&
		difficulty != models.DifficultyHard {
		_ = conn.WriteJSON(map[string]string{"type": "error", "message": "invalid difficulty"})
		return
	}

	// Создаем игру с ботом
	playerSymbol := m.gameManager.CreateBotGame(nickname, difficulty)
	botName := fmt.Sprintf("Bot_%s", difficulty)

	ctx := context.Background()
	if err := m.redis.Incr(ctx, "active_games").Err(); err != nil {
		logger.Warn("failed to increment active_games:", err)
	}

	// Отправляем подтверждение игроку
	_ = conn.WriteJSON(map[string]interface{}{
		"type":       "match_found",
		"symbol":     playerSymbol,
		"opponent":   botName,
		"isBot":      true,
		"difficulty": string(difficulty),
	})

	// Если бот ходит первым, делаем его ход
	if playerSymbol == "O" {
		time.AfterFunc(500*time.Millisecond, func() {
			m.makeBotMove(nickname)
		})
	}
}

func (m *WSManager) makeBotMove(nickname string) {
	game, ok := m.gameManager.GetGame(nickname)
	if !ok || !game.IsBotGame || game.IsFinished {
		return
	}

	botService := services.NewBotService()
	cell := botService.GetBotMove(game.Board, game.BotDifficulty, game.BotSymbol)
	if cell == -1 {
		return
	}

	// Определяем имя бота (тот игрок, чей символ == BotSymbol)
	botName := game.PlayerX
	if game.BotSymbol == "O" {
		// если бот O, то бот находится в PlayerO
		botName = game.PlayerO
	}

	// Выполняем ход как бот (передаём имя бота)
	moveMsg, resultMsg, err := m.gameManager.HandleMove(botName, cell)
	if err != nil {
		logger.Warn("Bot move error:", err)
		return
	}

	// Отправляем ответ игроку (и потенциально всем участникам игры)
	// Используем общую sendToGame, чтобы рассылка была единообразной.
	m.sendToGame(botName, moveMsg)
	if resultMsg != nil {
		m.sendToGame(botName, resultMsg)
		game.IsFinished = true
	} else {
		// проверяем ничью
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
			m.sendToGame(botName, drawMsg)
			game.IsFinished = true
		}
	}
}
