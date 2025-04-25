package http

import (
	"tictactoe/internal/services"
	wsManager "tictactoe/internal/ws"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(sessionService *services.SessionService) *gin.Engine {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5500"}, // python3 -m http.server 5500
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}))

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
