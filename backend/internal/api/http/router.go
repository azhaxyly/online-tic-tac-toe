package http

import (
	"io"
	"os"

	"tictactoe/internal/api/http/handlers"
	"tictactoe/internal/api/ws"
	"tictactoe/internal/services"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func NewRouter(sessionService *services.SessionService, leaderboardService *services.LeaderboardService) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	gin.DefaultWriter = io.Discard

	router := gin.New()

	allowedOrigin := os.Getenv("GIN_CORS_ALLOW_ORIGIN")
	if allowedOrigin == "" {
		allowedOrigin = "http://localhost:8080" // Для локальной разработки
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{allowedOrigin}, // Используем переменную
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Content-Type"},
		AllowCredentials: true,
	}))

	// Создаем middleware
	authMiddleware := AuthMiddleware(sessionService.RDB) // <-- НАШ MIDDLEWARE

	manager := ws.NewManager(sessionService.RDB, sessionService.Store)
	statsHandler := handlers.NewStatsHandler(sessionService.RDB)
	sessionHandler := handlers.NewSessionHandler(sessionService, sessionService.RDB)
	profileHandler := handlers.NewProfileHandler(sessionService.Store)
	leaderboardHandler := handlers.NewLeaderboardHandler(leaderboardService)
	shopService := services.NewShopService(sessionService.Store)
	shopHandler := handlers.NewShopHandler(shopService)

	// Защищенный WebSocket
	router.GET("/ws", authMiddleware, func(c *gin.Context) {
		// Получаем nickname из контекста, установленного middleware
		nickname, _ := c.Get("nickname")

		// Передаем nickname в HandleConnection
		manager.HandleConnection(c.Writer, c.Request, nickname.(string))
	})

	api := router.Group("/api")
	{
		api.POST("/register", sessionHandler.Register)
		api.POST("/login", sessionHandler.Login)
		api.POST("/logout", sessionHandler.Logout)

		api.GET("/stats", statsHandler.GetStats)
		api.GET("/leaderboard", leaderboardHandler.GetLeaderboard)

		api.GET("/nickname", authMiddleware, sessionHandler.GetNickname)
		api.GET("/profile-stats", authMiddleware, profileHandler.GetProfileStats)
		api.GET("/profile/:nickname", profileHandler.GetUserProfileByNickname)

		shop := api.Group("/shop")
		shop.Use(authMiddleware)
		{
			shop.GET("", shopHandler.GetShopInfo)
			shop.POST("/buy", shopHandler.BuyItem)
			shop.POST("/equip", shopHandler.EquipItem)
			shop.POST("/ad-reward", shopHandler.WatchAd)
		}
	}

	return router
}
