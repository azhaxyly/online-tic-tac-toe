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

	// Создаем middleware
	authMiddleware := AuthMiddleware(sessionService.RDB) // <-- НАШ MIDDLEWARE

	manager := ws.NewManager(sessionService.RDB)
	statsHandler := handlers.NewStatsHandler(sessionService.RDB)
	sessionHandler := handlers.NewSessionHandler(sessionService, sessionService.RDB)
	profileHandler := handlers.NewProfileHandler(sessionService.RDB)

	// Защищенный WebSocket
	router.GET("/ws", authMiddleware, func(c *gin.Context) {
		// Получаем nickname из контекста, установленного middleware
		nickname, _ := c.Get("nickname")

		// Передаем nickname в HandleConnection
		manager.HandleConnection(c.Writer, c.Request, nickname.(string))
	})

	api := router.Group("/api")
	{
		// ПУБЛИЧНЫЕ маршруты для аутентификации
		api.POST("/register", sessionHandler.Register)
		api.POST("/login", sessionHandler.Login)
		api.POST("/logout", sessionHandler.Logout)

		// Общая статистика (оставим публичной)
		api.GET("/stats", statsHandler.GetStats)

		// ЗАЩИЩЕННЫЕ маршруты
		// Проверка текущей сессии
		api.GET("/nickname", authMiddleware, sessionHandler.GetNickname)
		// Статистика профиля
		api.GET("/profile-stats", authMiddleware, profileHandler.GetProfileStats)
	}

	return router
}
