package handlers

import (
	"net/http"
	"tictactoe/internal/services"

	"github.com/gin-gonic/gin"
)

type ShopHandler struct {
	Service *services.ShopService
}

func NewShopHandler(service *services.ShopService) *ShopHandler {
	return &ShopHandler{Service: service}
}

func (h *ShopHandler) GetShopInfo(c *gin.Context) {
	nickname, exists := c.Get("nickname")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	info, err := h.Service.GetShopInfo(nickname.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, info)
}

func (h *ShopHandler) BuyItem(c *gin.Context) {
	nickname, exists := c.Get("nickname")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		ItemID string `json:"item_id"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.Service.BuyItem(nickname.(string), req.ItemID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *ShopHandler) WatchAd(c *gin.Context) {
	nickname, exists := c.Get("nickname")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	if err := h.Service.WatchAd(nickname.(string)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "rewarded", "amount": 50})
}

func (h *ShopHandler) EquipItem(c *gin.Context) {
	nickname, exists := c.Get("nickname")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		ItemID string `json:"item_id"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := h.Service.EquipItem(nickname.(string), req.ItemID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "equipped"})
}
