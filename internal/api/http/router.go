package http

import (
	"tictactoe/internal/services"

	"github.com/gin-gonic/gin"
)

func NewRouter(sessionService *services.SessionService) *gin.Engine {
	router := gin.Default()

	sessionHandler := NewSessionHandler(sessionService)

	api := router.Group("/api")
	{
		api.GET("/nickname", sessionHandler.GetNickname)
	}

	return router
}
