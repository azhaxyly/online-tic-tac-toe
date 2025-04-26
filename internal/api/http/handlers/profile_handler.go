package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type ProfileHandler struct {
	RDB *redis.Client
}

func NewProfileHandler(rdb *redis.Client) *ProfileHandler {
	return &ProfileHandler{
		RDB: rdb,
	}
}

func (h *ProfileHandler) GetProfileStats(c *gin.Context) {
	sessionIDCookie, err := c.Request.Cookie("session_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized (no session_id)"})
		return
	}
	sessionID := sessionIDCookie.Value

	ctx := context.Background()
	nickname, err := h.RDB.Get(ctx, "session:"+sessionID).Result()
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized (session not found)"})
		return
	}

	wins, _ := h.RDB.Get(ctx, "wins:"+nickname).Int()
	losses, _ := h.RDB.Get(ctx, "losses:"+nickname).Int()
	draws, _ := h.RDB.Get(ctx, "draws:"+nickname).Int()

	c.JSON(http.StatusOK, gin.H{
		"wins":   wins,
		"losses": losses,
		"draws":  draws,
	})
}
