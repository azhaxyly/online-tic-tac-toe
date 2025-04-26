package http

import (
	"tictactoe/internal/api/http/handlers"
	"tictactoe/internal/api/ws"
	"tictactoe/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(sessionService *services.SessionService) *gin.Engine {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"}, // python3 -m http.server 5500
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}))

	manager := ws.NewManager(sessionService.RDB)

	router.Static("/static", "./public")

	router.GET("/", func(c *gin.Context) {
		c.File("./public/test_game_full.html")
	})

	statsHandler := handlers.NewStatsHandler(sessionService.RDB)
	sessionHandler := handlers.NewSessionHandler(sessionService)
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
