package router

import (
	"context"
	"encoding/json"
	"log/slog"
	"math-top/internal/cache"
	"math-top/internal/config"
	"math-top/internal/handler"
	"math-top/internal/middleware"
	"math-top/internal/model"
	"math-top/internal/response"
	"math-top/internal/service"
	"math-top/internal/ws"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func NewEngine() *gin.Engine {
	if config.GlobalConfig == nil {
		config.LoadConfig()
	}
	gin.SetMode(config.GlobalConfig.App.Mode)

	r := gin.New()
	// 客户端 IP 信任链：gin 默认信任所有代理（0.0.0.0/0），攻击者伪造 X-Forwarded-For
	// 即可无限绕过限流（登录爆破、评论刷屏、注册轰炸）并污染 UV 统计。
	// 实测（2026-06）：固定 XFF 第 11 次登录返回 429；轮换 XFF 连续 15 次全部放行。
	// 这里只信任 Docker 桥接网段（nginx 反代所在网络）：nginx 的 XFF 会被采纳且
	// 从右往左取第一个非可信 IP；本地/公网直连的 XFF 一律忽略，按真实 RemoteAddr 限流。
	if err := r.SetTrustedProxies([]string{"172.16.0.0/12", "192.168.0.0/16", "10.0.0.0/8"}); err != nil {
		panic(err)
	}
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	// 请求体大小限制：JSON 1MB / multipart 64MB，防止无上限请求体造成内存 DoS
	r.Use(middleware.BodyLimitMiddleware())

	uploadBase := config.GlobalConfig.Storage.UploadDir
	if uploadBase == "" {
		uploadBase = "./storage/uploads"
	}
	// 公共上传目录逐个显式挂载，不再整根挂载 /uploads：
	// 1) 避免把 /uploads/chat 聊天室文件一并公开；
	// 2) gin 的路由树不允许 /uploads/chat/*filepath 与 /uploads/*filepath 同时注册
	//    （会启动 panic），所以“影子路由覆盖”方案不可用，改为按目录白名单挂载。
	publicDirs := []string{"avatars", "covers", "h5_unified_light", "resources"}
	for _, dir := range publicDirs {
		r.Static("/uploads/"+dir, filepath.Join(uploadBase, dir))
	}

	db := config.ConnectMysql()
	rdb := config.InitRedis()
	c := cache.New(rdb)

	// 聊天室文件访问控制：/uploads/chat 挂 JWT 鉴权静态路由。
	// 浏览器 <img>/<a> 无法带 Authorization 头，前端通过认证 fetch + blob URL 加载。
	chatStatic := r.Group("/uploads/chat")
	chatStatic.Use(middleware.JWTAuthMiddleware(rdb))
	chatStatic.Static("", filepath.Join(uploadBase, "chat"))

	// 聊天室 WebSocket Hub（单实例内存版；多实例时在此接 Redis Pub/Sub）
	hub := ws.NewHub()
	go hub.Run()
	broadcastFn := func(event string, data interface{}) {
		switch event {
		case ws.TypeMessage:
			hub.BroadcastMessage(data)
		case ws.TypeDelete:
			if ids, ok := data.([]uint); ok {
				hub.BroadcastDelete(ids)
			}
		case ws.TypePresence:
			if count, ok := data.(int64); ok {
				hub.BroadcastPresence(count)
			}
		}
	}

	// 通知实时推送：给指定用户的所有在线连接发送红点刷新事件（离线用户无连接、自然跳过）
	notifyFn := func(userIDs []uint) {
		payload, err := json.Marshal(ws.Envelope{Type: ws.TypeNotification, Data: map[string]interface{}{"kind": "notification"}})
		if err != nil {
			return
		}
		for _, uid := range userIDs {
			hub.SendToUser(uid, payload)
		}
	}

	if err := db.AutoMigrate(
		&model.User{},
		&model.Event{},
		&model.News{},
		&model.Resource{},
		&model.Favorite{},
		&model.Showcase{},
		&model.Admin{},
		&model.DownloadLog{},
		&model.Comment{},
		&model.CommentLike{},
		&model.ChatMessage{},
		&model.ChatPresence{},
		&model.DailyMetric{},
		&model.EventRegistration{},
		&model.Notification{},
		&model.NotificationBatch{},
	); err != nil {
		slog.Error("数据库迁移失败", "err", err)
		panic(err)
	}

	adminInitSvc := service.NewAdminService(db, rdb)
	if err := adminInitSvc.EnsureDefaultAdmin(); err != nil {
		slog.Error("创建默认管理员失败，服务终止", "err", err)
		panic(err)
	}

	// 每日指标归档常驻任务：不依赖管理员访问后台，防止 Redis 日指标 key 过期导致数据丢失
	go func() {
		// 启动后先补档一次（把昨天的数据落库）
		if err := adminInitSvc.SyncDailyMetrics(context.Background()); err != nil {
			slog.Warn("启动时同步每日指标失败", "err", err)
		}
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for range ticker.C {
			if err := adminInitSvc.SyncDailyMetrics(context.Background()); err != nil {
				slog.Warn("定时同步每日指标失败", "err", err)
			}
		}
	}()

	// 注入初始测试数据：仅当显式设置 SEED_DEMO_DATA=true 时执行，生产环境不设置
	if os.Getenv("SEED_DEMO_DATA") == "true" {
		seedSvc := service.NewSeedService(db)
		seedSvc.AutoSeed()
	}

	registerPublicRoutes(r, db, rdb, c)
	registerAuthRoutes(r, db, rdb, c, hub, broadcastFn, notifyFn)
	registerAdminRoutes(r, db, rdb, c, broadcastFn, notifyFn)

	// 通知实时推送 WebSocket：与聊天室共用 Hub、同子协议鉴权，但不触发聊天室在线状态
	notifyWSHandler := handler.NewNotifyWSHandler(hub, rdb)
	notifyWsLimit := middleware.RateLimitMiddleware(rdb, 30, time.Minute)
	r.Group("/api/v1").GET("/ws/notify", notifyWsLimit, notifyWSHandler.Handle)

	return r
}

func Addr() string {
	return ":" + config.GlobalConfig.App.Port
}

func registerPublicRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client, c *cache.Cache) {
	r.GET("/", func(c *gin.Context) {
		response.Success(c, gin.H{"message": "pong"})
	})
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, gin.H{"status": "ok"})
	})

	userService := service.NewUserService(db, rdb)
	homepageService := service.NewHomepageService(db, c)
	eventService := service.NewEventService(db)
	eventRegService := service.NewEventRegistrationService(db, rdb)
	newsService := service.NewNewsService(db)
	resourceService := service.NewResourceService(db)
	showcaseService := service.NewShowcaseService(db)
	commentService := service.NewCommentService(db)
	activityService := service.NewActivityService(rdb)

	userHandler := handler.NewUserHandler(userService)
	homepageHandler := handler.NewHomepageHandler(homepageService)
	eventHandler := handler.NewEventHandler(eventService, eventRegService)
	newsHandler := handler.NewNewsHandler(newsService)
	resourceHandler := handler.NewResourceHandler(resourceService)
	showcaseHandler := handler.NewShowcaseHandler(showcaseService)
	commentHandler := handler.NewCommentHandler(commentService)
	activityHandler := handler.NewActivityHandler(activityService)

	public := r.Group("/api/v1")
	{
		authLimit := middleware.RateLimitMiddleware(rdb, 10, time.Minute)
		trackLimit := middleware.RateLimitMiddleware(rdb, 60, time.Minute)
		public.POST("/auth/register", authLimit, userHandler.Register)
		public.POST("/auth/login", authLimit, userHandler.Login)
		public.POST("/auth/logout", userHandler.Logout)
		public.POST("/auth/forgot-password", authLimit, userHandler.ForgotPassword)
		public.POST("/auth/reset-password", authLimit, userHandler.ResetPassword)

		public.POST("/active/track", trackLimit, activityHandler.Track)

		public.GET("/home", homepageHandler.Get)

		public.GET("/events", eventHandler.List)
		public.GET("/events/:id", eventHandler.Detail)

		public.GET("/news", newsHandler.List)
		public.GET("/news/:id", newsHandler.Detail)

		public.GET("/resources", resourceHandler.List)
		public.GET("/resources/:id", resourceHandler.Detail)
		public.GET("/resources/download/:id", resourceHandler.Download)

		public.GET("/showcases", showcaseHandler.List)
		public.GET("/showcases/:id", showcaseHandler.Detail)

		public.GET("/comments", commentHandler.ListByTarget)
	}
}

func registerAuthRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client, c *cache.Cache, hub *ws.Hub, broadcastFn func(string, interface{}), notifyFn func([]uint)) {
	userService := service.NewUserService(db, rdb)
	memberService := service.NewMemberService(db)
	favoriteService := service.NewFavoriteService(db)
	resourceService := service.NewResourceService(db)
	commentService := service.NewCommentService(db)
	avatarService := service.NewAvatarService(db)
	chatService := service.NewChatService(db, c)
	chatService.SetBroadcast(broadcastFn)
	eventRegService := service.NewEventRegistrationService(db, rdb)
	eventRegService.SetNotifier(notifyFn)
	notificationService := service.NewNotificationService(db)

	userHandler := handler.NewUserHandler(userService)
	memberHandler := handler.NewMemberHandler(memberService)
	favoriteHandler := handler.NewFavoriteHandler(favoriteService)
	resourceHandler := handler.NewResourceHandler(resourceService)
	commentHandler := handler.NewCommentHandler(commentService)
	avatarHandler := handler.NewAvatarHandler(avatarService)
	chatHandler := handler.NewChatHandler(chatService)
	eventRegHandler := handler.NewEventRegistrationHandler(eventRegService)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	auth := r.Group("/api/v1")
	auth.Use(middleware.JWTAuthMiddleware(rdb))
	commentLimit := middleware.RateLimitMiddleware(rdb, 10, time.Minute)
	chatLimit := middleware.RateLimitMiddleware(rdb, 30, time.Minute)
	uploadLimit := middleware.RateLimitMiddleware(rdb, 20, time.Hour)
	{
		auth.POST("/auth/change-password", userHandler.ChangePassword)

		auth.GET("/profile", memberHandler.GetProfile)
		auth.PUT("/profile", memberHandler.UpdateProfile)
		auth.POST("/user/avatar", uploadLimit, avatarHandler.Upload)

		auth.POST("/member/favorites", favoriteHandler.AddFavorite)
		auth.DELETE("/member/favorites/:id", favoriteHandler.Remove)
		auth.GET("/member/favorites", favoriteHandler.ListFavorites)

		auth.GET("/member/downloads", resourceHandler.MyDownloads)

		auth.POST("/comments", commentLimit, commentHandler.Create)
		auth.DELETE("/comments/:id", commentHandler.Delete)
		auth.POST("/comments/:id/like", commentHandler.ToggleLike)

		auth.POST("/chat/join", chatHandler.Join)
		auth.GET("/chat/messages", chatHandler.ListMessages)
		auth.POST("/chat/messages", chatLimit, chatHandler.SendText)
		auth.POST("/chat/messages/file", uploadLimit, chatHandler.SendFile)
		auth.POST("/chat/leave", chatHandler.Leave)
		auth.DELETE("/chat/messages/:id", chatHandler.Delete)

		// 活动报名
		auth.POST("/member/events/:id/register", eventRegHandler.Register)
		auth.DELETE("/member/events/:id/register", eventRegHandler.Cancel)
		auth.GET("/member/events/registrations", eventRegHandler.MyRegistrations)

		// 消息通知（个人中心收件箱）
		auth.GET("/member/notifications", notificationHandler.List)
		auth.GET("/member/notifications/unread-count", notificationHandler.UnreadCount)
		auth.POST("/member/notifications/:id/read", notificationHandler.MarkRead)
		auth.POST("/member/notifications/read-all", notificationHandler.MarkAllRead)
	}

	// WebSocket 聊天室：不走普通 JWT 中间件，握手内部自鉴权（子协议携带 token）
	wsLimit := middleware.RateLimitMiddleware(rdb, 30, time.Minute)
	chatWSHandler := handler.NewChatWSHandler(chatService, hub, rdb)
	r.Group("/api/v1").GET("/chat/ws", wsLimit, chatWSHandler.Handle)
}

func registerAdminRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client, c *cache.Cache, broadcastFn func(string, interface{}), notifyFn func([]uint)) {
	adminService := service.NewAdminService(db, rdb)
	adminEventService := service.NewAdminEventService(db)
	adminNewsService := service.NewAdminNewsService(db)
	adminResourceService := service.NewAdminResourceService(db)
	adminShowcaseService := service.NewAdminShowcaseService(db)
	commentService := service.NewCommentService(db)
	adminUserService := service.NewAdminUserService(db)
	homepageService := service.NewHomepageService(db, c)
	resourceService := service.NewResourceService(db)
	chatService := service.NewChatService(db, c)
	chatService.SetBroadcast(broadcastFn)
	eventRegService := service.NewEventRegistrationService(db, rdb)
	notificationService := service.NewNotificationService(db)
	notificationService.SetNotifier(notifyFn)

	adminHandler := handler.NewAdminHandler(adminService)
	adminEventHandler := handler.NewAdminEventHandler(adminEventService)
	adminNewsHandler := handler.NewAdminNewsHandler(adminNewsService)
	adminResourceHandler := handler.NewAdminResourceHandler(adminResourceService)
	adminShowcaseHandler := handler.NewAdminShowcaseHandler(adminShowcaseService)
	adminUserHandler := handler.NewAdminUserHandler(adminUserService)
	adminCommentHandler := handler.NewAdminCommentHandler(commentService)
	adminChatHandler := handler.NewAdminChatHandler(chatService)
	resourceHandler := handler.NewResourceHandler(resourceService)
	eventRegHandler := handler.NewEventRegistrationHandler(eventRegService)
	notificationHandler := handler.NewNotificationHandler(notificationService)

	admin := r.Group("/api/v1/admin")
	adminLimit := middleware.RateLimitMiddleware(rdb, 20, time.Minute)
	admin.POST("/auth/login", adminLimit, adminHandler.Login)
	admin.POST("/auth/logout", adminHandler.Logout)

	auth := admin.Group("")
	auth.Use(middleware.AdminJWTAuthMiddleware(rdb))
	auth.Use(middleware.AdminAuthMiddleware())
	{
		auth.PUT("/auth/password", adminHandler.ChangePassword)
		auth.GET("/dashboard", adminHandler.Dashboard)
		auth.POST("/homepage/invalidate", func(c *gin.Context) {
			homepageService.Invalidate()
			response.Success(c, nil)
		})

		auth.GET("/events", adminEventHandler.List)
		auth.GET("/events/:id", adminEventHandler.Detail)
		auth.POST("/events", adminEventHandler.Create)
		auth.PUT("/events/:id", adminEventHandler.Update)
		auth.DELETE("/events/:id", adminEventHandler.Delete)
		auth.PATCH("/events/:id/feature", adminEventHandler.ToggleFeature)
		auth.PATCH("/events/:id/expired", adminEventHandler.SetExpired)

		// 活动报名管理（名单/签到/移除 + 汇总）
		auth.GET("/events/registration-summary", eventRegHandler.AdminSummary)
		auth.GET("/events/:id/registrations", eventRegHandler.AdminList)
		auth.POST("/events/:id/registrations/:uid/checkin", eventRegHandler.AdminCheckin)
		auth.POST("/events/:id/registrations/:uid/uncheckin", eventRegHandler.AdminCancelCheckin)
		auth.DELETE("/events/:id/registrations/:uid", eventRegHandler.AdminRemove)

		auth.GET("/news", adminNewsHandler.List)
		auth.GET("/news/:id", adminNewsHandler.Detail)
		auth.POST("/news", adminNewsHandler.Create)
		auth.PUT("/news/:id", adminNewsHandler.Update)
		auth.DELETE("/news/:id", adminNewsHandler.Delete)

		auth.GET("/resources", adminResourceHandler.List)
		auth.GET("/resources/:id", adminResourceHandler.Detail)
		auth.POST("/resources", resourceHandler.Upload)
		auth.PUT("/resources/:id", adminResourceHandler.Update)
		auth.DELETE("/resources/:id", adminResourceHandler.Delete)

		auth.GET("/showcases", adminShowcaseHandler.List)
		auth.GET("/showcases/:id", adminShowcaseHandler.Detail)
		auth.POST("/showcases", adminShowcaseHandler.Create)
		auth.PUT("/showcases/:id", adminShowcaseHandler.Update)
		auth.DELETE("/showcases/:id", adminShowcaseHandler.Delete)

		auth.GET("/users", adminUserHandler.List)
		auth.GET("/users/export", adminUserHandler.Export)
		auth.PATCH("/users/:id/status", adminUserHandler.SetStatus)
		auth.POST("/users/:id/reset-password", adminUserHandler.ResetPassword)
		auth.DELETE("/users/:id", adminUserHandler.Delete)
		auth.POST("/users/batch-status", adminUserHandler.BatchSetStatus)
		auth.POST("/users/batch-reset-password", adminUserHandler.BatchResetPassword)
		auth.POST("/users/batch-delete", adminUserHandler.BatchDelete)

		auth.GET("/comments", adminCommentHandler.List)
		auth.DELETE("/comments/:id", adminCommentHandler.Delete)
		auth.PATCH("/comments/:id/status", adminCommentHandler.SetStatus)

		auth.GET("/chat/messages", adminChatHandler.List)
		auth.DELETE("/chat/messages/:id", adminChatHandler.Delete)

		// 通知管理（单个/部门/全部发送 + 发送记录）
		auth.POST("/notifications", notificationHandler.Send)
		auth.GET("/notifications", notificationHandler.Batches)
	}
}
