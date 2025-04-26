package handlers

import (
	"fmt"
	"net/http"

	"tictactoe/internal/logger"
	"tictactoe/internal/services"
	"tictactoe/internal/utils"

	"github.com/gin-gonic/gin"
)

type SessionHandler struct {
	session *services.SessionService
}

func NewSessionHandler(s *services.SessionService) *SessionHandler {
	return &SessionHandler{session: s}
}

func (h *SessionHandler) GetNickname(c *gin.Context) {
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

	logger.Info(fmt.Sprintf("User %s connected with session ID %s", user.Nickname, sessionID))
	c.JSON(http.StatusOK, gin.H{"nickname": user.Nickname})
}
