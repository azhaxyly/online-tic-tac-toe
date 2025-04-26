package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type StatsHandler struct {
	RDB *redis.Client
}

func NewStatsHandler(rdb *redis.Client) *StatsHandler {
	return &StatsHandler{
		RDB: rdb,
	}
}

func (h *StatsHandler) GetStats(c *gin.Context) {
	ctx := context.Background()

	online, err1 := h.RDB.Get(ctx, "online_users").Int()
	if err1 == redis.Nil {
		online = 0
	} else if err1 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get online_users"})
		return
	}

	activeGames, err2 := h.RDB.Get(ctx, "active_games").Int()
	if err2 == redis.Nil {
		activeGames = 0
	} else if err2 != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get active_games"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"online":       online,
		"active_games": activeGames,
	})
}
