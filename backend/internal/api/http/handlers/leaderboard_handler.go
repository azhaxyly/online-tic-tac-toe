package handlers

import (
	"net/http"
	"tictactoe/internal/services"

	"github.com/gin-gonic/gin"
)

type LeaderboardHandler struct {
	service *services.LeaderboardService
}

func NewLeaderboardHandler(service *services.LeaderboardService) *LeaderboardHandler {
	return &LeaderboardHandler{
		service: service,
	}
}

func (h *LeaderboardHandler) GetLeaderboard(c *gin.Context) {
	stats, err := h.service.GetLeaderboard()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch leaderboard"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
