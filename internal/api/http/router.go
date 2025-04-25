package http

import (
	"tictactoe/internal/services"
	wsManager "tictactoe/internal/ws"

	"github.com/gin-gonic/gin"
)

func NewRouter(sessionService *services.SessionService) *gin.Engine {
	router := gin.Default()

	manager := wsManager.NewManager(sessionService.RDB)

	// WS route
	router.GET("/ws", func(c *gin.Context) {
		manager.HandleConnection(c.Writer, c.Request)
	})

	// HTTP API
	api := router.Group("/api")
	{
		api.GET("/nickname", NewSessionHandler(sessionService).GetNickname)
	}

	return router
}
