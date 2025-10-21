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
	// 1. Получаем nickname из контекста, который был установлен
	//    в AuthMiddleware.
	nicknameVal, exists := c.Get("nickname")
	if !exists {
		// Эта проверка - на всякий случай. Если middleware настроен правильно,
		// "nickname" должен всегда существовать.
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	nickname := nicknameVal.(string)
	ctx := context.Background()

	// 2. Получаем статистику из Redis по ключам
	//    Мы используем .Int() и игнорируем ошибку ( _ ).
	//    Если ключ (например, "wins:nickname") не существует (redis.Nil),
	//    .Int() вернет 0, что нам и нужно.
	wins, _ := h.RDB.Get(ctx, "wins:"+nickname).Int()
	losses, _ := h.RDB.Get(ctx, "losses:"+nickname).Int()
	draws, _ := h.RDB.Get(ctx, "draws:"+nickname).Int()

	// 3. Отправляем JSON-ответ
	c.JSON(http.StatusOK, gin.H{
		"wins":   wins,
		"losses": losses,
		"draws":  draws,
	})
}
