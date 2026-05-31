package router

import (
	"net/http"

	appconfig "github.com/time/card/backend/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/time/card/backend/internal/handler"
	"github.com/time/card/backend/internal/middleware"
	"github.com/time/card/backend/internal/service"
	"gorm.io/gorm"
)

func New(cfg *appconfig.Config, db *gorm.DB, rdb *redis.Client) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)

	authService := service.NewAuthService(db, &cfg.Auth)
	douDizhuService := service.NewDouDizhuService()
	roomService := service.NewDouDizhuRoomService()
	zhajinhuaService := service.NewZhajinhuaService()
	zhajinhuaRoomService := service.NewZhajinhuaRoomService()

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), corsMiddleware())

	health := &handler.HealthHandler{DB: db, Redis: rdb}
	r.GET("/health", health.Check)
	r.GET("/api/health", health.Check)

	authHandler := &handler.AuthHandler{Auth: authService}
	r.POST("/api/auth/login", authHandler.Login)

	gameHandler := &handler.GameHandler{DouDizhu: douDizhuService}
	roomHandler := &handler.RoomHandler{Rooms: roomService, DouDizhu: douDizhuService}
	zhHandler := &handler.ZhajinhuaHandler{Games: zhajinhuaService, Rooms: zhajinhuaRoomService}

	api := r.Group("/api")
	api.Use(middleware.AuthRequired(authService))
	{
		api.GET("/auth/me", authHandler.Me)
		api.GET("/games/catalog", gameHandler.Catalog)
		api.POST("/games/doudizhu/start", gameHandler.StartDouDizhu)
		api.POST("/games/doudizhu/rooms/join", roomHandler.Join)
		api.GET("/games/doudizhu/rooms/:roomId", roomHandler.Get)
		api.POST("/games/doudizhu/rooms/:roomId/leave", roomHandler.Leave)
		api.POST("/games/doudizhu/rooms/:roomId/ready", roomHandler.Ready)
		api.POST("/games/doudizhu/rooms/:roomId/next", roomHandler.Next)
		api.GET("/games/doudizhu/:gameId", gameHandler.GetDouDizhuState)
		api.POST("/games/doudizhu/:gameId/call", gameHandler.CallDouDizhu)
		api.POST("/games/doudizhu/:gameId/play", gameHandler.PlayDouDizhu)
		api.POST("/games/doudizhu/:gameId/pass", gameHandler.PassDouDizhu)
		api.GET("/games/doudizhu/:gameId/hint", gameHandler.HintDouDizhu)
		api.POST("/games/doudizhu/:gameId/tick", gameHandler.TickDouDizhu)

		api.POST("/games/zhajinhua/start", zhHandler.Start)
		api.POST("/games/zhajinhua/rooms/join", zhHandler.JoinRoom)
		api.GET("/games/zhajinhua/rooms/:roomId", zhHandler.GetRoom)
		api.POST("/games/zhajinhua/rooms/:roomId/leave", zhHandler.LeaveRoom)
		api.POST("/games/zhajinhua/rooms/:roomId/ready", zhHandler.ReadyRoom)
		api.POST("/games/zhajinhua/rooms/:roomId/start", zhHandler.StartRoom)
		api.POST("/games/zhajinhua/rooms/:roomId/next", zhHandler.ReadyNext)
		api.GET("/games/zhajinhua/:gameId", zhHandler.GetState)
		api.POST("/games/zhajinhua/:gameId/look", zhHandler.Look)
		api.POST("/games/zhajinhua/:gameId/fold", zhHandler.Fold)
		api.POST("/games/zhajinhua/:gameId/follow", zhHandler.Follow)
		api.POST("/games/zhajinhua/:gameId/raise", zhHandler.Raise)
		api.POST("/games/zhajinhua/:gameId/compare", zhHandler.Compare)
		api.POST("/games/zhajinhua/:gameId/tick", zhHandler.Tick)
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
