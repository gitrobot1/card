package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	appconfig "github.com/time/card/backend/internal/config"
	"github.com/time/card/backend/internal/handler"
	"github.com/time/card/backend/internal/middleware"
	"github.com/time/card/backend/internal/service"
	cardws "github.com/time/card/backend/internal/ws"
	"gorm.io/gorm"
)

func New(cfg *appconfig.Config, db *gorm.DB, rdb *redis.Client) *gin.Engine {
	gin.SetMode(cfg.Server.Mode)

	authService := service.NewAuthService(db, &cfg.Auth)
	douDizhuService := service.NewDouDizhuService()
	roomService := service.NewDouDizhuRoomService()
	zhajinhuaService := service.NewZhajinhuaService()
	zhajinhuaRoomService := service.NewZhajinhuaRoomService()
	unoService := service.NewUnoService()
	unoRoomService := service.NewUnoRoomService()
	douniuService := service.NewDouNiuService()
	douniuRoomService := service.NewDouNiuRoomService()
	douniuHub := cardws.NewDouNiuHub()
	yuzhoushaService := service.NewYuzhoushaService()
	yuzhoushaRoomService := service.NewYuzhoushaRoomService()
	yuzhoushaHub := cardws.NewYuzhoushaHub()

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
	unoHandler := &handler.UnoHandler{Games: unoService, Rooms: unoRoomService}
	dnHandler := &handler.DouNiuHandler{Games: douniuService, Rooms: douniuRoomService, Hub: douniuHub}
	dnWSHandler := &handler.DouNiuWSHandler{Auth: authService, Games: douniuService, Rooms: douniuRoomService, Hub: douniuHub}
	yzsHandler := &handler.YuzhoushaHandler{Games: yuzhoushaService, Rooms: yuzhoushaRoomService, Hub: yuzhoushaHub}
	yzsWSHandler := &handler.YuzhoushaWSHandler{Auth: authService, Games: yuzhoushaService, Rooms: yuzhoushaRoomService, Hub: yuzhoushaHub}

	r.GET("/ws/douniu/rooms/:roomId", dnWSHandler.Room)
	r.GET("/ws/douniu/games/:gameId", dnWSHandler.Game)
	r.GET("/ws/yuzhousha/rooms/:roomId", yzsWSHandler.Room)
	r.GET("/ws/yuzhousha/games/:gameId", yzsWSHandler.Game)

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

		api.POST("/games/uno/start", unoHandler.Start)
		api.POST("/games/uno/rooms/join", unoHandler.JoinRoom)
		api.GET("/games/uno/rooms/:roomId", unoHandler.GetRoom)
		api.POST("/games/uno/rooms/:roomId/leave", unoHandler.LeaveRoom)
		api.POST("/games/uno/rooms/:roomId/ready", unoHandler.ReadyRoom)
		api.POST("/games/uno/rooms/:roomId/start", unoHandler.StartRoom)
		api.POST("/games/uno/rooms/:roomId/next", unoHandler.ReadyNext)
		api.GET("/games/uno/:gameId", unoHandler.GetState)
		api.POST("/games/uno/:gameId/play", unoHandler.Play)
		api.POST("/games/uno/:gameId/draw", unoHandler.Draw)
		api.POST("/games/uno/:gameId/vote-end", unoHandler.VoteEnd)
		api.POST("/games/uno/:gameId/roll-first", unoHandler.RollFirst)
		api.POST("/games/uno/:gameId/tick", unoHandler.Tick)

		api.POST("/games/douniu/start", dnHandler.Start)
		api.POST("/games/douniu/rooms/join", dnHandler.JoinRoom)
		api.GET("/games/douniu/rooms/:roomId", dnHandler.GetRoom)
		api.POST("/games/douniu/rooms/:roomId/leave", dnHandler.LeaveRoom)
		api.POST("/games/douniu/rooms/:roomId/ready", dnHandler.ReadyRoom)
		api.POST("/games/douniu/rooms/:roomId/start", dnHandler.StartRoom)
		api.POST("/games/douniu/rooms/:roomId/next", dnHandler.ReadyNext)
		api.GET("/games/douniu/:gameId", dnHandler.GetState)
		api.POST("/games/douniu/:gameId/grab", dnHandler.GrabBanker)
		api.POST("/games/douniu/:gameId/bet", dnHandler.PlaceBet)
		api.POST("/games/douniu/:gameId/tick", dnHandler.Tick)

		api.GET("/games/yuzhousha/modes", yzsHandler.Modes)
		api.GET("/games/yuzhousha/packs", yzsHandler.Packs)
		api.GET("/games/yuzhousha/heroes", yzsHandler.Heroes)
		api.POST("/games/yuzhousha/start", yzsHandler.Start)
		api.POST("/games/yuzhousha/rooms/join", yzsHandler.JoinRoom)
		api.GET("/games/yuzhousha/rooms/:roomId", yzsHandler.GetRoom)
		api.POST("/games/yuzhousha/rooms/:roomId/leave", yzsHandler.LeaveRoom)
		api.POST("/games/yuzhousha/rooms/:roomId/hero", yzsHandler.SetHeroRoom)
		api.POST("/games/yuzhousha/rooms/:roomId/ready", yzsHandler.ReadyRoom)
		api.POST("/games/yuzhousha/rooms/:roomId/start", yzsHandler.StartRoom)
		api.POST("/games/yuzhousha/rooms/:roomId/next", yzsHandler.ReadyNextRoom)
		api.GET("/games/yuzhousha/:gameId", yzsHandler.GetState)
		api.POST("/games/yuzhousha/:gameId/skill", yzsHandler.UseSkill)
		api.POST("/games/yuzhousha/:gameId/play", yzsHandler.PlayCard)
		api.POST("/games/yuzhousha/:gameId/shan", yzsHandler.RespondShan)
		api.POST("/games/yuzhousha/:gameId/respond", yzsHandler.RespondCard)
		api.POST("/games/yuzhousha/:gameId/pass", yzsHandler.PassResponse)
		api.POST("/games/yuzhousha/:gameId/pass-all-wuxiek", yzsHandler.PassAllWuxiek)
		api.POST("/games/yuzhousha/:gameId/bagua", yzsHandler.BaguaJudge)
		api.POST("/games/yuzhousha/:gameId/end", yzsHandler.EndPlay)
		api.POST("/games/yuzhousha/:gameId/discard", yzsHandler.DiscardCard)
		api.POST("/games/yuzhousha/:gameId/weapon/discard", yzsHandler.RespondWeaponDiscard)
		api.POST("/games/yuzhousha/:gameId/prepare/pass", yzsHandler.PassPrepare)
		api.POST("/games/yuzhousha/:gameId/draw/pass", yzsHandler.PassDraw)
		api.POST("/games/yuzhousha/:gameId/peek-deck", yzsHandler.FinishPeekDeck)
		api.POST("/games/yuzhousha/:gameId/guanxing", yzsHandler.FinishGuanxing)
		api.POST("/games/yuzhousha/:gameId/tick", yzsHandler.Tick)
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
