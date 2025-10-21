package http

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func AuthMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized (no session)"})
			return
		}

		ctx := context.Background()
		nickname, err := rdb.Get(ctx, "session:"+sessionID).Result()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized (invalid session)"})
			return
		}

		// Сессия валидна. Обновим ее время жизни
		rdb.Expire(ctx, "session:"+sessionID, 24*time.Hour)
		// Также обновим "online" ключ
		rdb.Set(ctx, "online:"+sessionID, 1, 3*time.Minute)

		// Сохраняем nickname в контексте Gin для
		// последующих обработчиков
		c.Set("nickname", nickname)

		c.Next()
	}
}
