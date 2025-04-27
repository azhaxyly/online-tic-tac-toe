package handlers

import (
	"fmt"
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

func (h *SessionHandler) GetNickname(c *gin.Context) {
	ctx := c.Request.Context()
	sessionID, err := c.Cookie("session_id")
	if err != nil || sessionID == "" {
		sessionID = utils.GenerateSessionID()
		c.SetCookie("session_id", sessionID, 3600*24, "/", "", false, true)
	}

	user, err := h.session.GetOrCreateUser(sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	err = h.RDB.Set(ctx, "online:"+sessionID, 1, 3*time.Minute).Err()
	if err != nil {
		logger.Warn("Failed to update online status:", err)
	}

	logger.Info(fmt.Sprintf("User %s connected with session ID %s", user.Nickname, sessionID))
	c.JSON(http.StatusOK, gin.H{"nickname": user.Nickname})
}
