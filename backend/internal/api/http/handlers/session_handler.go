package handlers

import (
	"net/http"
	"time"

	"tictactoe/internal/logger"
	"tictactoe/internal/services"
	"tictactoe/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type SessionHandler struct {
	session *services.SessionService
	RDB     *redis.Client
}

func NewSessionHandler(s *services.SessionService, RDB *redis.Client) *SessionHandler {
	return &SessionHandler{
		session: s,
		RDB:     RDB,
	}
}

// Структуры для JSON-запросов
type RegisterRequest struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

func (h *SessionHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := h.session.Register(req.Nickname, req.Password)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"nickname": user.Nickname})
}

func (h *SessionHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	user, err := h.session.Login(req.Nickname, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Успешный вход! Создаем сессию
	ctx := c.Request.Context()
	sessionID := utils.GenerateSessionID()

	// Сохраняем сессию в Redis
	err = h.RDB.Set(ctx, "session:"+sessionID, user.Nickname, 24*time.Hour).Err()
	if err != nil {
		logger.Error("Failed to set session in Redis:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Устанавливаем cookie
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie("session_id", sessionID, 3600*24, "/", "", true, true)
	logger.Info("User logged in:", user.Nickname)
	c.JSON(http.StatusOK, gin.H{"nickname": user.Nickname})
}

func (h *SessionHandler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err == nil && sessionID != "" {
		ctx := c.Request.Context()
		h.RDB.Del(ctx, "session:"+sessionID)
		h.RDB.Del(ctx, "online:"+sessionID) // Также удаляем из онлайна
	}

	// Очищаем cookie
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie("session_id", "", -1, "/", "", true, true)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *SessionHandler) GetNickname(c *gin.Context) {
	// Мы доверяем middleware. Если он нас пропустил, значит,
	// никнейм уже есть в контексте.
	nickname, exists := c.Get("nickname")
	if !exists {
		// Эта ситуация не должна произойти, если middleware настроен верно
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"nickname": nickname})
}
