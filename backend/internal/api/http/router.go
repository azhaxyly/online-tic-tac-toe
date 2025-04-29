package http

import (
	"io"
	"tictactoe/internal/api/http/handlers"
	"tictactoe/internal/api/ws"
	"tictactoe/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(sessionService *services.SessionService) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	router := gin.New()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}))

	manager := ws.NewManager(sessionService.RDB)

	statsHandler := handlers.NewStatsHandler(sessionService.RDB)
	sessionHandler := handlers.NewSessionHandler(sessionService, sessionService.RDB)
	profileHandler := handlers.NewProfileHandler(sessionService.RDB)

	router.GET("/ws", func(c *gin.Context) {
		manager.HandleConnection(c.Writer, c.Request)
	})

	api := router.Group("/api")
	{
		api.GET("/nickname", sessionHandler.GetNickname)
		api.GET("/stats", statsHandler.GetStats)
		api.GET("/profile-stats", profileHandler.GetProfileStats)
	}

	return router
}
