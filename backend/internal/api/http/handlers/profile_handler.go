package handlers

import (
	"net/http"

	"tictactoe/internal/store"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	UserStore *store.UserStore
}

func NewProfileHandler(userStore *store.UserStore) *ProfileHandler {
	return &ProfileHandler{
		UserStore: userStore,
	}
}

// GetProfileStats returns the authenticated user's profile
func (h *ProfileHandler) GetProfileStats(c *gin.Context) {
	nicknameVal, exists := c.Get("nickname")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	nickname := nicknameVal.(string)
	user, err := h.UserStore.GetUserProfile(nickname)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nickname":   user.Nickname,
		"wins":       user.Wins,
		"losses":     user.Losses,
		"draws":      user.Draws,
		"elo_rating": user.EloRating,
	})
}

// GetUserProfileByNickname returns any user's public profile by nickname
func (h *ProfileHandler) GetUserProfileByNickname(c *gin.Context) {
	nickname := c.Param("nickname")
	if nickname == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nickname required"})
		return
	}

	user, err := h.UserStore.GetUserProfile(nickname)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"nickname":   user.Nickname,
		"wins":       user.Wins,
		"losses":     user.Losses,
		"draws":      user.Draws,
		"elo_rating": user.EloRating,
	})
}
